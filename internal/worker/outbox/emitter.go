package outbox

//event-workerのヘルパーを置く。
import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

//パッケージ横断しないからexport関数ではない
// jsonlRow は JSON Lines の1行分。record の必要フィールドだけ出力する。
/*emitted_at/lease_owner/lease_until/next_attempt_at/claimed_atは
配送・実行制御のためのメタデータであって、イベントの意味そのものではないからjsonに含めない。
S3は時系列で検索しない要件。
S3はclaimed_atベース
*/

// encoding/json がアクセスできるのは export（大文字）フィールドだけやから大文字にする。DBも同様。

// JSONL は機械用やけど人間も読める。parquetはさらに機械用
type jsonlRow struct {
	ID            string          `json:"id"`
	UserID        string          `json:"user_id"`
	TaskID        string          `json:"task_id"`
	RequestID     string          `json:"request_id"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	SchemaVersion uint32          `json:"schema_version"`
	Payload       json.RawMessage `json:"payload"`
}

// buildJSONLines は ClaimedEvent のスライスから JSON Lines バイト列を生成する。
func buildJSONLines(records []ClaimedEvent) ([]byte, error) {
	//普通encodeはjson.NewEncoder(w).Encode(v)みたいな感じやけどこれはすぐに記述。
	//今回は一旦Encode → buffer に置いてから次にbuffer → reader。
	//encoderはio.Writer(interface)型やったら可能やから簡単にbufferとかresponsewriterとか切り替えれる。
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	//デフォルトではjsonに< >があると\u003cにエスケープされる。(XSS対策)
	//今回はhtmlの中に埋め込まないから人間が読みやすいようにescapeなしにする。enc.SetEscapeHTML(false)
	for i := range records {
		row := jsonlRow{
			ID:            records[i].ID,
			UserID:        records[i].UserID,
			TaskID:        records[i].TaskID,
			RequestID:     records[i].RequestID,
			EventType:     records[i].EventType,
			OccurredAt:    records[i].OccurredAt,
			SchemaVersion: records[i].SchemaVersion,
			Payload:       records[i].Payload,
		}
		if err := enc.Encode(row); err != nil {
			return nil, fmt.Errorf("encode jsonl row id=%s: %w", records[i].ID, err)
		}
	}
	//jsonlをbytesに詰め込んで返す
	return buf.Bytes(), nil
}

// manifest は S3 にアップロードする成功確定マーカー。
type manifest struct {
	BatchID   string    `json:"batch_id"`  //バッチを一意に識別するID
	DataKey   string    `json:"data_key"`  //どこのデータ。
	EventIDs  []string  `json:"event_ids"` //これは任意やけど監査しやすいから作る。
	Count     int       `json:"count"`     //データの数。
	CreatedAt time.Time `json:"created_at"`
}

//なんでhashIDは適当にuuid.Newとかulidで作らずにわざわざhashにしてる。
//uuidとかulidは毎回違う値を返す。今回は同様の入力では同様の値が欲しい(manifestで検索)。
//同様の入力バイト列 → 必ず同様ののハッシュ値になる。
// batchID は決定的ハッシュで生成する。同じ入力 → 同じキー → 上書きで冪等。
/*同じイベントなら複数回実行しても同様のS3キーになるようにしている。*/
func batchID(eventIDs []string, claimedAt time.Time, schemaVersion uint32) string {
	//eventid軍が同じやけど順番違うから別っていうパターンを弾くために並び替える。
	sorted := make([]string, len(eventIDs))
	//並び替えたいけどもとのIDsは崩したくないからsortedにコピーする。
	copy(sorted, eventIDs)
	//copyはbuilt-in関数。len関数とかと同様。
	sort.Strings(sorted) //並べ替え。

	h := sha256.New() //hash計算機を作る(interface返す)。
	/*type Hash interface {
	    Write(p []byte) (n int, err error)
	    Sum(b []byte) []byte
	    Reset()
	    Size() int
	    BlockSize() int
	}*/
	//これはセキュリティ用途じゃないからinternal/security使わない。
	// そもそも実装も違う。普通にsha256使う。別に差し替えもしない。"crypto/sha256"依存許容。
	for _, id := range sorted { //idがインデックスではなく値。
		h.Write([]byte(id)) //writeはbyte引数。writeで計算機に値を詰める->sumで取り出す。
		//TODO:["ab","c"] と ["a","bc"]考慮
		//返り値無視してる。厳密にやるならばerrチェックする。
	}
	h.Write([]byte(claimedAt.Truncate(time.Hour).Format(time.RFC3339)))
	//truncateは時間を丸める。2026-02-13 15:42:10.123->2026-02-13 15:00:00
	//eventID変わらんけど日時が違う場合同様のhashになること防ぐ、そのためにclaimedat使う。
	binary.Write(h, binary.BigEndian, schemaVersion)
	//encoding/binaryパッケージ。整数など(今回はschemaVersionをhに記述)を決まったバイト順(今回はBigEndian)でバイト列にして書くためのユーティリティ。
	//BigEndianでバイト方式を固定。

	return hex.EncodeToString(h.Sum(nil))[:16] //返す
}

// 04d->4桁でない部分は0で埋める。
// S3はファイル単位ではなくkey(raw/<prefix>/year=YYYY/month=MM/day=DD/hour=HH/<bid>.jsonl)で保存先を決める。ここでは保存先を返す。
// bidはbatchID。バッチを一意に識別する。
// prefix はログの種類 / サービス名 / ドメイン。todo、authとか。
/*raw/todo/...
raw/payment/...
raw/auth/...になって見やすい
マイクロサービスではかなり便利*/
func s3DataKey(prefix string, claimedAt time.Time, bid string) string { //string返す
	return fmt.Sprintf("raw/%s/year=%04d/month=%02d/day=%02d/hour=%02d/%s.jsonl", //fmt.Sprintfはstring返す。fmt.Sprintf忘れがち。
		prefix,
		claimedAt.Year(), claimedAt.Month(), claimedAt.Day(), claimedAt.Hour(),
		//claimedateは2026-02-13 15:42:10.123こういう形式。
		bid,
	)
}

//そもそもこれ同一関数にしてjsoknとmanifest分岐させたらコード少なくなるけど分かりにくい。これ大事な感覚

func s3ManifestKey(prefix string, claimedAt time.Time, bid string) string {
	return fmt.Sprintf("raw/%s/year=%04d/month=%02d/day=%02d/hour=%02d/%s.manifest.json",
		prefix,
		claimedAt.Year(), claimedAt.Month(), claimedAt.Day(), claimedAt.Hour(),
		bid,
	)
}

//recordからeventID格納する。使う場面多い。
// batchID(collectIDs(records), claimedAt, schemaVersion)
//repo.MarkEmitted(ctx, ids, owner, now)

func collectIDs(records []ClaimedEvent) []string {
	ids := make([]string, len(records))
	for i := range records {
		ids[i] = records[i].ID
	}
	return ids
}

// trimByByteLimit は records を payload バイト合計が maxBytes 以内に収まるよう切る。
// claimした後削る関数
func trimByByteLimit(records []ClaimedEvent, maxBytes int64) []ClaimedEvent {

	var total int64 //int64適切
	for i := range records {
		total += int64(len(records[i].Payload))
		if total > maxBytes {
			return records[:i] //これ便利
		}
	}
	return records
}

// manifestはファイル完成しましたという宣言->中身0ではない
/*jsonlはrecords []ClaimedEvent->1件ずつ row に変換
 manifestは既に完成系が入力
claim する->JSONL を作る->S3 に upload->成功した->じゃあ manifest 作る->manifest を upload->DB を MarkEmitted
*/

func buildManifestJSON(m manifest) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	//jsonlは一行一レコード。
	//func (enc *Encoder) SetIndent(prefix, indent string)　何も付与しない(prefixなし)改行してスペース"  "(indent)
	/*{"batch_id":"abc","count":10}から
		{
	  "batch_id": "abc",
	  "count": 10
	}
	へ変換。manifestは人間が見る
	*/
	if err := enc.Encode(m); err != nil {
		return nil, fmt.Errorf("encode manifest: %w", err)
	}
	return buf.Bytes(), nil
}

//"todo/" "/todo" "/todo/"で与えられる場合ある。それをnormalizeする。
//TrimSuffixで最後の/消してTrimPrefixで最初の/消す。

func normalizePrefix(prefix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(prefix, "/"), "/")
}

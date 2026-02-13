package taskeventrepo

//送る準備の段階。claim.go
import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Claim は未処理イベントを FOR UPDATE SKIP LOCKED でロックし取得する。
// 呼び出し側がトランザクション内で実行する前提。
func (r *repository) Claim(
	ctx context.Context,
	limit int,
	now time.Time, //repoでtime定義ではなくclockから受け取る->テストしやすい。
) ([]TaskEventRecord, error) {
	//ここは全カラム取ってない。大量イベント取得ではスキャンコストが大きな差になる。select段階で状態制御系カラムはスキャンしない。

	const q = `
SELECT
  id, user_id, task_id, request_id, event_type,
  occurred_at, schema_version, payload, attempt_count
FROM task_events
WHERE emitted_at IS NULL
  AND next_attempt_at <= ?
  AND (lease_until IS NULL OR lease_until < ?)
ORDER BY occurred_at ASC
LIMIT ?
FOR UPDATE SKIP LOCKED;
`
	//そもそもheartbeatしてるなら絶対処理時間内に終わるくない。全体が詰まってる可能性。->別ワーカーに奪われる。
	//発生順に扱うと便利
	//FOR UPDATE SKIP LOCKEDはとった行をロック、ロックされてるならばスキップ。
	//DB側で同一行編集を弾く、それでも完全ではない、実行側で冪等を完全に保証する(2段階)
	//そのためのlease。そもそもfor updateはトランザクション中だけやけど作業はトランザクション終わっても続くことある->lease。
	var records []TaskEventRecord
	if err := r.q.SelectContext(ctx, &records, q, now, now, limit); err != nil { //まずctx与える。
		return nil, fmt.Errorf("taskevent claim select: %w", err)
	}
	return records, nil
}

// SetLease は指定されたイベントIDのリースを設定する。
// Claim で取得した行のうち、Go側でバイト上限チェック後に確定したIDだけを対象とする。
func (r *repository) SetLease(
	ctx context.Context,
	ids []string,
	leaseOwner string,
	leaseDuration time.Duration,
	now time.Time,
) error {
	if len(ids) == 0 {
		return nil
	}
	const base = `
UPDATE task_events
SET lease_owner = ?,
    lease_until = ?,
    claimed_at = ?
WHERE id IN (?);
`
	//ここはwhereで条件考えてない。selectで保証してそれを同一txで使ってる。ここでもう一回whereでガードするのは冗長。
	//claimed_atはselectで設定したほうがよくねと思ったけど違う。selectした後超過分は削られるからここで設定が適切。
	//そもそもロックで見つけた時刻ではなく担当が確定した時刻が定義
	//ここでattemptcountは扱わずに失敗のカウントとしてattempt_countを考える。
	leaseUntil := now.Add(leaseDuration)
	//時間計算はusecaseでやってleaseUntilを引数にしてないのはなんで。claimed_atのnowとlease_untilで基準としたnowがそろってることが分かりやすい。

	query, args, err := sqlx.In(base, leaseOwner, leaseUntil, now, ids)
	//sqlx.Inはidsを展開する。sqlx.Inは?プレースホルダーを扱う。
	//query:=IN(?)が展開されたSQLargs:=idsを展開した引数
	//ここでsqlxに依存するのは許容。最低限usecaseはinfraがsqlxを使ってることに関与しない。
	if err != nil {
		return fmt.Errorf("taskevent setlease expand in: %w", err)
	}
	query = r.q.Rebind(query)
	//sqlx.Rebindはsqlx.Inで?になったのをドライバーに合わせて変換する。今回はmysql使ってるから実際はいらんけど切り替えを見据える。
	//query = sqlx.Rebind(sqlx.QUESTION, query)これにするとハードコード。
	//だからqueryerexecerにRebind(query string) stringこれ定義してDBによって勝手に切り替わるようにする。

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("taskevent setlease execute: %w", err)
	}
	return nil
}

// ExtendLeaseはCAS付きでリースを延長する。
//Compare-And-Swap。SET lease_until=? WHERE id IN (?) AND lease_owner=?はただのswap
// SET lease_until=? WHERE id IN (?) AND lease_owner=? AND lease_until=?は状態に基づいて比べたswap。leaseownerは比べてるわけではなくただの識別。
// RowsAffected=0 の場合はリースが別ワーカーに奪われたことを意味する。
/*Worker A が Claim→SetLease して処理開始（lease_until = T+60s）
A が何らかで詰まって heartbeat/extend が止まる（GC停止、ネット詰まり、goroutine停止、プロセス生存はしてる 等）
時間が経って lease_until を過ぎる
Worker B が Claim でその行を拾って SetLease（lease_owner=B, lease_until=T+120s）
その後 A が復帰して「延長しよ」と ExtendLease を投げる
→ でも DB上の owner/lease_until はもう違う
→ RowsAffected=0
延長はselect経由しないからロック効かない。*/
func (r *repository) ExtendLease(
	ctx context.Context,
	ids []string,
	leaseOwner string,
	currentLeaseUntil time.Time,
	leaseDuration time.Duration,
	now time.Time,
) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	//この後変換されるからqではなくbaseが適切。
	const base = `
UPDATE task_events
SET lease_until = ?
WHERE id IN (?)
  AND lease_owner = ?
  AND lease_until = ?;
` //lease_ownerだけでなくlease_untilもチェック->CAS
	newLeaseUntil := now.Add(leaseDuration)

	query, args, err := sqlx.In(base, newLeaseUntil, ids, leaseOwner, currentLeaseUntil)
	if err != nil {
		return 0, fmt.Errorf("taskevent extendlease expand in: %w", err)
		//自分が定義したエラーはerrで返す。infra層のエラー、外部のエラーはwrapして返す。エラー起こった箇所を伝えることが目的。
	}
	query = r.q.Rebind(query)

	res, err := r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("taskevent extendlease execute: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("taskevent extendlease rowsaffected: %w", err)
	}
	/*Claimの FOR UPDATE：Tx中だけ
	lease_owner/lease_until：Tx後の論理ロックだが、期限切れたら回収される
	->rowsaffected 三段階*/
	return affected, nil //これaffected返して==0はここで定義してない
	//いつもの楽観ロックとは違い起こりうるケース。やからinfraエラーにしない。
}

// ReleaseLease はリースをクリアする（失敗時の後処理用）。ここでattempt_countは操作しない。
func (r *repository) ReleaseLease(
	ctx context.Context,
	ids []string,
	leaseOwner string,
) error {
	if len(ids) == 0 {
		return nil
	}
	const base = `
UPDATE task_events
SET lease_owner = NULL,
    lease_until = NULL
WHERE id IN (?)
  AND lease_owner = ?;
`
	query, args, err := sqlx.In(base, ids, leaseOwner)
	if err != nil {
		return fmt.Errorf("taskevent releaselease expand in: %w", err)
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("taskevent releaselease execute: %w", err)
	}
	return nil
}

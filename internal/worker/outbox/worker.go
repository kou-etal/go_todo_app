package outbox

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/logger"
	//loggerに寄せたけどslogはstdlib。外部ライブラリと違って依存と呼ぶほどのものかは議論の余地がある。
	//requestidを自動で取得してくれるのは便利。
)

type Worker struct {
	repo     OutboxRepo
	uploader ObjectUploader //usecaseはs3の具体には関与しない。
	cfg      Config
	ownerID  string
	s3Prefix string
	logger   logger.Logger
}

func NewWorker(repo OutboxRepo, uploader ObjectUploader, cfg Config, logger logger.Logger) *Worker {
	return &Worker{
		repo:     repo,
		uploader: uploader,
		cfg:      cfg,
		ownerID:  uuid.New().String(), //ここでownerIDが一意になるからCASが保険になる。
		s3Prefix: normalizePrefix("task-events"),
		logger:   logger,
	}
}

// Run はメインループ。ctx がキャンセルされるまで繰り返す。
func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info(ctx, "outbox worker started", logger.String("owner", w.ownerID))
	for {
		//selectはswitchとは違う。チャンネル専用の制御構文
		//selectはループごとにctxが終わってないか判断する。defaultがないと終わるまで待ち続ける。defaultは終わるのを待たない。
		select {
		case <-ctx.Done():
			w.logger.Info(ctx, "outbox worker shutting down")
			return nil
		default:
		}

		processed, err := w.processOnce(ctx)
		if err != nil {
			w.logger.Error(ctx, "outbox worker process error", err)
		}

		// イベントがあった場合は即ループ(ifに入らない)、なければsleep
		//これでdefaultが短いスパンで実行され続けることを避ける。
		if !processed {
			select {
			//ただのtime.Sleep(w.cfg.IdleSleep)はsleepの間にキャンセル来てもわからない。
			//time.Afterでn秒後に信号送る->cancellable wait。
			case <-ctx.Done():
				return nil
			case <-time.After(w.cfg.IdleSleep):
			}
		}
	}
}

// processOnce は1回の claim → emit サイクルを実行する。
// イベントを処理した場合 true を返す。
func (w *Worker) processOnce(ctx context.Context) (bool, error) {
	//ここでの時間はtodo-apiのusecaseと違ってビジネスロジック(deadline判定とか)に関与しない。
	//普段todo-apiのusecaseをclockerのutils使ってるけどそれはテストしやすいから。今回はビジネスロジック内からテストなし。
	//time.Now()使うこと多い。
	now := time.Now()

	// 1. Claim: 候補をロック
	records, err := w.repo.Claim(ctx, w.cfg.ChunkMaxRows, now)
	if err != nil {
		return false, fmt.Errorf("claim: %w", err)
	} //分類する意味がない。if errors.Is(err, ErrDeadlock) { ... }。if errors.Is(err, ErrTimeout) { ... }とかいらん。
	// 失敗したという事実だけでいい。どこで失敗したかわかれば十分。重要な考え方。
	//worker/infraエラーはエンジニアが見る。domain/usecaseエラーはユーザーが見る。
	if len(records) == 0 {
		return false, nil
	}

	// 2. Go側でバイトチェック
	records = trimByByteLimit(records, w.cfg.ChunkMaxBytes)
	//ここを真レコードと扱う。SetLeaseはrecord変えないからわざわざとこでrecord取得しなくていい。
	if len(records) == 0 {
		return false, nil
	}

	ids := collectIDs(records)
	//ここで時間更新しないとleaseが短くなる。
	now = time.Now()

	// 3. 確定分だけリース設定
	if err := w.repo.SetLease(ctx, ids, w.ownerID, w.cfg.LeaseDuration, now); err != nil {
		return false, fmt.Errorf("set lease: %w", err)
	}
	leaseUntil := now.Add(w.cfg.LeaseDuration) //heartbeatはleaseUntil使う。
	claimedAt := now                           //リース設定をclaim基準判定。ここでclaimedat定義。

	// 4. Heartbeat goroutine 開始
	//heartbeatはgo routine
	hbCtx, hbCancel := context.WithCancel(ctx) //標準記法
	defer hbCancel()                           //終わらせないとメモリリーク。
	//終了はctx伝播、deferは後の掃除作業
	//goroutine内でsource作ってないまたは 作っても自動で終わる/GCで回収されるだけ なら、deferなくてもいい。
	//でもgo routineはdeferで片づける習慣。
	go w.heartbeatLoop(hbCtx, ids, leaseUntil) //これできるんや。レシーバーの中でレシーバー使う

	// 5. S3 にアップロード
	emitErr := w.emitToS3(ctx, records, ids, claimedAt)
	hbCancel()

	if emitErr != nil {
		w.handleFailure(ctx, ids, records, emitErr)
		return true, nil
	}

	// 6. 成功: emitted_at を設定
	if err := w.repo.MarkEmitted(ctx, ids, w.ownerID, time.Now()); err != nil {
		return true, fmt.Errorf("mark emitted: %w", err)
	}

	w.logger.Info(ctx, "emitted events", logger.Int("count", len(ids)))
	return true, nil
}

// emitToS3 はデータファイル + manifest を S3 にアップロードする。
func (w *Worker) emitToS3(
	ctx context.Context,
	records []ClaimedEvent,
	ids []string,
	claimedAt time.Time,
) error {

	bid := batchID(ids, claimedAt, 1)                //schemaversion1、ここはパーティションキーのclaimedat使う。
	dataKey := s3DataKey(w.s3Prefix, claimedAt, bid) //ここでclaimedat使ってるからパーティションキー
	manifestKey := s3ManifestKey(w.s3Prefix, claimedAt, bid)

	// manifest が既に存在するなら、前回アップロード成功済み → DB 更新だけ再実行
	//存在はmanifestkeyで確認。
	exists, err := w.uploader.Exists(ctx, manifestKey)
	if err != nil {
		return fmt.Errorf("check manifest exists: %w", err)
	}
	if exists {
		w.logger.Info(ctx, "manifest already exists, skipping upload", logger.String("key", manifestKey))
		//こういう普通に起こりうる場合はlogger.info使う。
		return nil
	}

	// データファイル (.jsonl) アップロード
	//これは記述してくれるわけではない。bufferを返すだけ。
	jsonl, err := buildJSONLines(records)
	//キーに含めたclaimedatはjsonlには含めない。
	if err != nil {
		return fmt.Errorf("build jsonl: %w", err)
	}
	if err := w.uploader.Upload(ctx, dataKey, bytes.NewReader(jsonl)); err != nil { //s3実装はinfra層。usecaseで定義しない。
		//ここで記述。 bytes.NewReader(jsonl)
		return fmt.Errorf("upload data: %w", err)
	}

	// Manifest アップロード（ここが成功確定）
	//manifestは人間が見たいデータ。CreatedAtが欲しい。
	m := manifest{
		BatchID:   bid,
		DataKey:   dataKey,
		EventIDs:  ids,
		Count:     len(ids),
		CreatedAt: time.Now(),
	}
	mJSON, err := buildManifestJSON(m)
	if err != nil {
		return fmt.Errorf("build manifest json: %w", err)
	}
	if err := w.uploader.Upload(ctx, manifestKey, bytes.NewReader(mJSON)); err != nil {
		return fmt.Errorf("upload manifest: %w", err)
	}
	return nil
	//ここでのerrはemitErrになる。
}

// handleFailure はアップロード失敗時の処理。max_attempt 超過で DLQ に移動。
func (w *Worker) handleFailure(
	ctx context.Context,
	ids []string,
	records []ClaimedEvent,
	emitErr error,
) {
	w.logger.Error(ctx, "emit failed", emitErr, logger.Int("count", len(ids)))

	// max_attempt チェック: records の中で最大の attempt_count を見る
	//schemaが INT UNSIGNEDの時、多くのドライバや生成ツールはuint32にマッピングする。
	//BIGINT UNSIGNEDの場合uint64。今回の場合uint32で十分。
	var maxCount uint32
	for i := range records {
		if records[i].AttemptCount >= maxCount {
			maxCount = records[i].AttemptCount
		} //これmax探索アルゴリズム
		//Go では普通に自分で回して最大値取るのが標準。ライブラリでmax求めないこと多い。浸透してないから読みにくい。
		//TODO:これattemptcount低いのにDLQが起こるから全部同一を保証する設計にする
	}

	// attempt_count は SetLease 時点の値。今回の失敗で +1 されるので +1 で比較。
	if maxCount+1 >= w.cfg.MaxAttempt {
		w.logger.Warn(ctx, "moving to DLQ", logger.Attr{Key: "ids", Value: ids}, logger.Int("attempts", int(maxCount)+1))
		if err := w.repo.MoveToDLQ(ctx, ids, emitErr.Error(), time.Now()); err != nil {
			//引数はerrorではなくstring。emitErr.Error()使う。
			w.logger.Error(ctx, "move to DLQ failed", err)
		}
		return
	}

	nextAt := NextAttemptAt(time.Now(), int(maxCount)+1, w.cfg.BackoffBase) //ここはキャストする。
	if err := w.repo.MarkRetry(ctx, ids, w.ownerID, nextAt); err != nil {
		w.logger.Error(ctx, "mark retry failed", err)
	}
}

// heartbeatLoop は定期的にリースを延長する。CAS で安全性を保証。
// これはgoroutineで動く
func (w *Worker) heartbeatLoop(ctx context.Context, ids []string, currentLeaseUntil time.Time) {
	ticker := time.NewTicker(w.cfg.HeartbeatInterval)
	//失敗した時のnextはjitter。heartbeatは固定interval。
	//ticker:=一定間隔で信号が届くチャンネルを作るもの。
	//defaultないから信号届くまで待つ。
	defer ticker.Stop()
	//ticker はタイマー資源を持つから、終了時に止めないと無駄に動き続ける（リーク）。

	for {
		//これは終わらないとループし続けるfor。正常に終わるか親がattempt_maxに到達して終了シグナルで終わる。
		// あるいはworker奪われる。エラー。
		select {
		case <-ctx.Done():
			return
			//これもcancellable wait
		case <-ticker.C:
			now := time.Now() //更新必須。時間使う->更新
			affected, err := w.repo.ExtendLease(
				ctx, ids, w.ownerID,
				currentLeaseUntil, w.cfg.LeaseDuration, now,
			)
			if err != nil {
				w.logger.Error(ctx, "heartbeat extend lease error", err)
				return
			}
			if affected == 0 { //worker奪われた場合。
				// これはtodo-apiのupdateと違って起こりうるからinfraでエラー返さずにusecaseでエラーなしで返す。
				w.logger.Warn(ctx, "heartbeat CAS mismatch, lease lost")
				return
			}
			currentLeaseUntil = now.Add(w.cfg.LeaseDuration)
			//これは意味ある。次のループまでにもう一回processOnceのnow.Add(w.cfg.LeaseDuration)通らない。
		}
	}
}

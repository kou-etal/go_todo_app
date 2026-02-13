package outbox

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

//ここは抽象を置くだけ。具体実装はinfraのeventとs3に任せる。

// clean architectureではusecaseがinfraをimportは不可。

// ClaimedEvent はワーカーが扱うイベントの表現。
// infra層のDB構造体(TaskEventRecord)をworkerに関与させないためのDTO。
type ClaimedEvent struct {
	ID            string
	UserID        string
	TaskID        string
	RequestID     string
	EventType     string
	OccurredAt    time.Time
	SchemaVersion uint32
	Payload       json.RawMessage
	AttemptCount  uint32
}

/*普段やったらdomain抽象作ってそのあとusecaseのdepsでdeps定義するところやけど
workerの場合ドメインルールがなくusecaseがメインやからdomainを無しでinfraを直接depsにまとめてtxに使うイメージ
もしworkerがドメインルール持つならdomain使う*/

// OutboxRepo はワーカーが必要とするリポジトリ操作の契約。
// infra/repository/event の *repository が満たす。
type OutboxRepo interface {
	Claim(ctx context.Context, limit int, now time.Time) ([]ClaimedEvent, error)
	SetLease(ctx context.Context, ids []string, leaseOwner string, leaseDuration time.Duration, now time.Time) error
	ReleaseLease(ctx context.Context, ids []string, leaseOwner string) error
	ExtendLease(ctx context.Context, ids []string, leaseOwner string, currentLeaseUntil time.Time, leaseDuration time.Duration, now time.Time) (int64, error)
	MarkEmitted(ctx context.Context, ids []string, leaseOwner string, now time.Time) error
	MarkRetry(ctx context.Context, ids []string, leaseOwner string, nextAttemptAt time.Time) error
	MoveToDLQ(ctx context.Context, ids []string, lastError string, now time.Time) error
}

// ObjectUploader は S3 互換ストレージへのアップロード抽象。
// infra/s3 の実装が満たす。
type ObjectUploader interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Exists(ctx context.Context, key string) (bool, error)
}

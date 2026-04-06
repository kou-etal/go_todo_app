package outbox

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

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

type OutboxRepo interface {
	Claim(ctx context.Context, limit int, now time.Time) ([]ClaimedEvent, error)
	SetLease(ctx context.Context, ids []string, leaseOwner string, leaseDuration time.Duration, now time.Time) error
	ReleaseLease(ctx context.Context, ids []string, leaseOwner string) error
	ExtendLease(ctx context.Context, ids []string, leaseOwner string, currentLeaseUntil time.Time, leaseDuration time.Duration, now time.Time) (int64, error)
	MarkEmitted(ctx context.Context, ids []string, leaseOwner string, now time.Time) error
	MarkRetry(ctx context.Context, ids []string, leaseOwner string, nextAttemptAt time.Time) error
	MoveToDLQ(ctx context.Context, ids []string, lastError string, now time.Time) error
	CountUnemitted(ctx context.Context) (int64, error)
}

type ObjectUploader interface {
	Upload(ctx context.Context, key string, body io.Reader) error
	Exists(ctx context.Context, key string) (bool, error)
}

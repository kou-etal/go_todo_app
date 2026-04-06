package taskeventrepo

//実際の送信ではなく送る準備の段階。
import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kou-etal/go_todo_app/internal/worker/outbox"
)

// Claim は未処理イベントを FOR UPDATE SKIP LOCKED でロックし取得する。
// 呼び出し側がトランザクション内で実行する前提。
func (r *repository) Claim(
	ctx context.Context,
	limit int,
	now time.Time,
) ([]outbox.ClaimedEvent, error) {

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

	var records []TaskEventRecord
	if err := r.q.SelectContext(ctx, &records, q, now, now, limit); err != nil {
		return nil, fmt.Errorf("taskevent claim select: %w", err)
	}
	return toClaimedEvents(records), nil
}

func toClaimedEvents(records []TaskEventRecord) []outbox.ClaimedEvent {
	events := make([]outbox.ClaimedEvent, len(records))
	for i, r := range records {
		events[i] = outbox.ClaimedEvent{
			ID:            r.ID,
			UserID:        r.UserID,
			TaskID:        r.TaskID,
			RequestID:     r.RequestID,
			EventType:     r.EventType,
			OccurredAt:    r.OccurredAt,
			SchemaVersion: r.SchemaVersion,
			Payload:       r.Payload,
			AttemptCount:  r.AttemptCount,
		}
	}
	return events
}

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

	leaseUntil := now.Add(leaseDuration)

	query, args, err := sqlx.In(base, leaseOwner, leaseUntil, now, ids)

	if err != nil {
		return fmt.Errorf("taskevent setlease expand in: %w", err)
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("taskevent setlease execute: %w", err)
	}
	return nil
}

// ExtendLeaseはCAS付きでリースを延長する。

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

	const base = `
UPDATE task_events
SET lease_until = ?
WHERE id IN (?)
  AND lease_owner = ?
  AND lease_until = ?;
`
	newLeaseUntil := now.Add(leaseDuration)

	query, args, err := sqlx.In(base, newLeaseUntil, ids, leaseOwner, currentLeaseUntil)
	if err != nil {
		return 0, fmt.Errorf("taskevent extendlease expand in: %w", err)
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
	
	return affected, nil
}


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

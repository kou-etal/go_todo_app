package taskeventrepo

//実際の送信ではなく送った後の段階。実際の送信はそもそもinfra/repositoryで定義しない。
//internal/worker/outboxで実際の送信定義。
import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// MarkEmitted は配信成功したイベントの emitted_at を設定し、リースをクリアする。
// 失敗のlease clearはclaim.goに置く。ここには正常系の状態遷移だけ置く
func (r *repository) MarkEmitted(
	ctx context.Context,
	ids []string,
	leaseOwner string,
	now time.Time,
) error {
	if len(ids) == 0 {
		return nil
	}
	const base = `
UPDATE task_events
SET emitted_at = ?,
    lease_owner = NULL,
    lease_until = NULL
WHERE id IN (?)
  AND lease_owner = ?;
`
	//claim->始まり。emit->終わり
	query, args, err := sqlx.In(base, now, ids, leaseOwner)
	if err != nil {
		return fmt.Errorf("taskevent markemitted expand in: %w", err)
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("taskevent markemitted execute: %w", err)
	}
	return nil
}

// MarkRetry は配信失敗したイベントの attempt_count をインクリメントし、
// next_attempt_at をバックオフ値に設定してリースをクリアする。
// これはclaim.goリース開放とは違う。正常な失敗。
func (r *repository) MarkRetry(
	ctx context.Context,
	ids []string,
	leaseOwner string,
	nextAttemptAt time.Time,
) error {
	if len(ids) == 0 {
		return nil
	}
	const base = `
UPDATE task_events
SET attempt_count = attempt_count + 1,
    next_attempt_at = ?,
    lease_owner = NULL,
    lease_until = NULL
WHERE id IN (?)
  AND lease_owner = ?;
`
	query, args, err := sqlx.In(base, nextAttemptAt, ids, leaseOwner)
	if err != nil {
		return fmt.Errorf("taskevent markretry expand in: %w", err)
	}
	query = r.q.Rebind(query)

	_, err = r.q.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("taskevent markretry execute: %w", err)
	}
	return nil
}

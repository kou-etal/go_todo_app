package taskeventrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// MoveToDLQ は max_attempt 超過したイベントを task_events_dlq に移動し、task_events から削除する。
// トランザクションで呼ばれる前提。
// task_events_dlq に INSERTとtask_events から DELETEは同一イベントサイクル->一つのrepoに記述。
// 同じドメインの同じ集約/同じライフサイクルの操作は一つのrepoで良い
// 別ドメインの永続化。例えばユーザー登録+メール送信は当然別ドメインやから分ける。
// それを同一repoにすることは可能やがそれはrepoがusecaseに関与してる。
func (r *repository) MoveToDLQ(
	ctx context.Context,
	ids []string,
	lastError string,
	now time.Time,
) error {
	if len(ids) == 0 {
		return nil
	}

	// 1. task_events_dlq に INSERT
	const insertBase = `
INSERT INTO task_events_dlq (
  id, user_id, task_id, request_id, event_type,
  occurred_at, schema_version, payload,
  attempt_count, last_error, dead_at
)
SELECT
  id, user_id, task_id, request_id, event_type,
  occurred_at, schema_version, payload,
  attempt_count, ?, ?
FROM task_events
WHERE id IN (?);
` //max件数を決めてるからWHERE id IN (?)可能。決めてなかったら件数大量で重くなる。
	insertQuery, insertArgs, err := sqlx.In(insertBase, lastError, now, ids)
	if err != nil {
		return fmt.Errorf("taskevent movetodlq insert expand in: %w", err)
	}
	insertQuery = r.q.Rebind(insertQuery)

	_, err = r.q.ExecContext(ctx, insertQuery, insertArgs...)
	if err != nil {
		return fmt.Errorf("taskevent movetodlq insert execute: %w", err)
	}

	// 2. task_events から DELETE
	const deleteBase = `
DELETE FROM task_events
WHERE id IN (?);
`
	deleteQuery, deleteArgs, err := sqlx.In(deleteBase, ids)
	if err != nil {
		return fmt.Errorf("taskevent movetodlq delete expand in: %w", err)
	}
	deleteQuery = r.q.Rebind(deleteQuery)

	_, err = r.q.ExecContext(ctx, deleteQuery, deleteArgs...)
	if err != nil {
		return fmt.Errorf("taskevent movetodlq delete execute: %w", err)
	}
	return nil
}

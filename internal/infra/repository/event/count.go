package taskeventrepo

import (
	"context"
	"fmt"
)

// CountUnemitted は未 emit のイベント数を返す。
// outbox worker の queue depth 計測用。低頻度で呼ばれる前提。
func (r *repository) CountUnemitted(ctx context.Context) (int64, error) {
	const q = `SELECT COUNT(*) FROM task_events WHERE emitted_at IS NULL`
	var count int64
	if err := r.q.GetContext(ctx, &count, q); err != nil {
		return 0, fmt.Errorf("taskevent count unemitted: %w", err)
	}
	return count, nil
}

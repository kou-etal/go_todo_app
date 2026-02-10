package taskeventrepo

import (
	"context"
	"fmt"

	dtaskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
)

func (r *repository) Insert(ctx context.Context, e *dtaskevent.TaskEvent) error {
	//emittedはnullで消費されたらmarkする。
	const q = `
INSERT INTO task_events (
  id, user_id, task_id, request_id,
  event_type, occurred_at,
  emitted_at, schema_version, payload
) VALUES (
  :id, :user_id, :task_id, :request_id,
  :event_type, :occurred_at,
  :emitted_at, :schema_version, :payload
);
`

	rec, err := toRecord(e)
	if err != nil {
		return err
	}

	_, err = r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("taskevent insert execute: %w", err)
	}
	return nil
}

package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *Repository) FindByID(
	ctx context.Context,
	id dtask.TaskID,
) (*dtask.Task, error) {

	const q = `
SELECT
  id,title,description,status,due_date,created_at,updated_at,version
FROM task
WHERE id = ?
LIMIT 1
`

	var rec TaskRecord
	if err := r.q.GetContext(ctx, &rec, q, id.Value()); err != nil {
		if isNotFound(err) {
			return nil, dtask.ErrNotFound
		} //norowsを吸収
		return nil, fmt.Errorf("taskrepo findbyid: %w", err)
	}

	t, err := RecordToEntity(&rec)
	if err != nil {
		return nil, fmt.Errorf("invalid task record id=%s: %w", rec.ID, err)
	}

	return t, nil
}

package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *repository) Store(ctx context.Context, t *dtask.Task) error {

	const q = `
INSERT INTO task (
  id, user_id, title, description, status, due_date,
  created_at, updated_at, version
) VALUES (
  :id, :user_id,:title, :description, :status, :due_date,
  :created_at, :updated_at, :version
);
`
	rec := EntityToRecord(t)

	_, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("taskrepo store execute: %w", err)
	}
	return nil
}
func (r *repository) Update(ctx context.Context, t *dtask.Task) error {
	const q = `
UPDATE task
SET
  title = :title,
  description = :description,
  status = :status,
  due_date = :due_date,
  updated_at = :updated_at,
  version = :next_version
WHERE
  id = :id
  AND user_id = :user_id
  AND version = :version;
`

	rec := EntityToRecord(t)
	params := map[string]any{
		"id":           rec.ID,
		"user_id":      rec.UserID,
		"title":        rec.Title,
		"description":  rec.Description,
		"status":       rec.Status,
		"due_date":     rec.DueDate,
		"updated_at":   rec.Updated,
		"version":      rec.Version,
		"next_version": rec.Version + 1,
	}

	res, err := r.q.NamedExecContext(ctx, q, params)
	if err != nil {
		return fmt.Errorf("taskrepo update execute: %w", err)
	}

	ra, err := res.RowsAffected()

	if err != nil {
		return fmt.Errorf("taskrepo update rowsaffected: %w", err)
	}
	if ra == 0 {
		return dtask.ErrConflict
	}
	return nil
}

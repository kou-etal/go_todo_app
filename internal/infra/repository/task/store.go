package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *Repository) Store(ctx context.Context, t *dtask.Task) error {
	//errorがないなら成功フラグ
	//queryは与える。commandは持ち込まない。
	const q = `
INSERT INTO task (
  id, title, description, status, due_date,
  created_at, updated_at, version
) VALUES (
  :id, :title, :description, :status, :due_date,
  :created_at, :updated_at, :version
);
` //ここは不変ゆえにconstでいい。

	rec := EntityToRecord(t)

	_, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("taskrepo store execute: %w", err)
	}
	return nil
}

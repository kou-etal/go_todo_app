package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *repository) Delete(ctx context.Context, id dtask.TaskID, version uint64) error {
	const q = `
DELETE FROM task
WHERE
  id = ?
  AND version = ?;
`
	res, err := r.q.ExecContext(ctx, q, id.Value(), version)

	if err != nil {
		return fmt.Errorf("taskrepo delete execute: %w", err)
	}
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("taskrepo delete rowsaffected: %w", err)
	}

	if ra == 0 {

		return dtask.ErrConflict
	} //TODO:ここちゃんとerrnotfoundとconflictに分ける。

	return nil
}

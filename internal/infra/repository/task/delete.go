package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

//論理削除or物理削除
/*UPDATE task
SET deleted_at = NOW()
WHERE id = ?で消えたように扱うのが論理削除*/
//論理削除はログが残るし復元可能
//今回はdeleted_at作らない物理削除
func (r *Repository) Delete(ctx context.Context, id dtask.TaskID, version uint64) error {
	const q = `
DELETE FROM task
WHERE
  id = ?
  AND version = ?;
`
	res, err := r.q.ExecContext(ctx, q, id.Value(), version)
	//今回は引数多くないからmappingしなくていい。ExecContext
	if err != nil {
		return fmt.Errorf("taskrepo delete execute: %w", err)
	} //repoでの撃ち落としなし
	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("taskrepo delete rowsaffected: %w", err)
	}

	if ra == 0 {
		//存在しないorVersion不一致
		return dtask.ErrConflict
	} //TODO:ここちゃんとerrnotfoundとconflictに分ける。
	//deleteではversionを増やさない

	return nil
}

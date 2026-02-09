package taskrepo

import (
	"context"
	"fmt"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *repository) Store(ctx context.Context, t *dtask.Task) error {
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
  AND version = :version;
` //versionの一致まで確かめる。楽観ロック

	rec := EntityToRecord(t)
	params := map[string]any{
		"id":           rec.ID,
		"title":        rec.Title,
		"description":  rec.Description,
		"status":       rec.Status,
		"due_date":     rec.DueDate,
		"updated_at":   rec.Updated,
		"version":      rec.Version,
		"next_version": rec.Version + 1,
	}
	//帰納的に+1を圧合うためのparams。+1の責務をdomainに寄せるならばrecをsqlに直接代入可能
	//いや違うわ。別にnext_versionにする意味ないか。普通にversion+1でもいいけど今回は全データsetで見た目整ってるからparam作ってる。
	//TODO:でもコードだるくなるし+1のほうがいいか
	//+1の責務をここで持ってることによりdomain側のversionは更新されない。ゆえに今の設計やと更新した後にリロード必須。

	res, err := r.q.NamedExecContext(ctx, q, params)
	if err != nil {
		return fmt.Errorf("taskrepo update execute: %w", err)
	}

	ra, err := res.RowsAffected()
	//RowsAffectedは何行更新したか返す。insertもupdateも一行。これが0ならば条件不一致。楽観ロック。
	//いや違うわ。0の時は楽観ロックor存在しない。
	// TODO:だから存在しない場合はErrNotFoundに分類してもいい。
	if err != nil {
		return fmt.Errorf("taskrepo update rowsaffected: %w", err)
	}
	if ra == 0 {
		return dtask.ErrConflict
	}
	return nil
}

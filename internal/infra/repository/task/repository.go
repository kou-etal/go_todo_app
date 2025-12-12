package taskrepo

import (
	"context"
	"errors"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.Queryer
}

var _ dtask.TaskRepository = (*Repository)(nil)

func NewRepository(q db.Queryer) *Repository {
	return &Repository{q: q}
}

func (r *Repository) ListAll(ctx context.Context) ([]*dtask.Task, error) {
	const sql = `
        SELECT
            id, title, description, status, dueDate,
            created, updated, version
        FROM task
        ORDER BY created DESC, id DESC;
    `
	//sqlが増えると別ファイルに切り出してもいい
	//created同一が存在するかもしれない。
	//var records []*TaskRecord 学習用メモ　これの場合毎回新しいヒープが発生、巨大レコード、加工や共有したい場合
	var records []TaskRecord
	if err := r.q.SelectContext(ctx, &records, sql); err != nil {
		return nil, errors.New("tmp")
	}

	tasks := make([]*dtask.Task, 0, len(records))
	//学習用メモ　capactity設定しないとキャパオーバーしたら新たに配列作られる。0、lenがベストプラクティス
	for i := range records {
		rec := &records[i]
		t, err := RecordToEntity(rec)
		if err != nil {
			return nil, errors.New("tmp")
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

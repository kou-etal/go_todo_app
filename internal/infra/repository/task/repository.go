package taskrepo

import (
	"context"
	"fmt"
	"strings"

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

func (r *Repository) List(ctx context.Context, q dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {
	//usecaseで決めてもいいが、repoでも最低限守る
	//limitはこっちで決めてクライアントに決めさせないほうがいいらしい
	if q.Limit <= 0 {
		q.Limit = 50
	}
	if q.Limit > 200 {
		q.Limit = 200
	}
	if q.Sort == "" {
		q.Sort = dtask.SortCreated
	}

	var sb strings.Builder
	//適当に6
	args := make([]any, 0, 6)

	sb.WriteString(`
SELECT
  id, title, description, status, dueDate,
  created, updated, version
FROM task
`)

	if q.Cursor != nil {
		switch q.Sort {
		case dtask.SortCreated:
			sb.WriteString("WHERE (created, id) < (?, ?)\n")
			//ORDER BY と WHERE のキー一致WHERE (created < ? OR (created = ? AND id < ?))よりもいい
			args = append(args, q.Cursor.Created, q.Cursor.ID.Value())
			//args = append(args, q.Cursor.Created.UTC(), q.Cursor.Created.UTC(), q.Cursor.ID.String())
			//domain repoの境界でutcにしたからここでutcはしない。すでに保証されている。
		//現在はDueIsNull=trueになることがないが仕様変更を見据えてtrueの場合のsql処理もかいてる。
		//duedate is nullの時のduedate比較をしていない。ifで分けてる。
		case dtask.SortDueDate:
			if !q.Cursor.DueIsNull {
				//WHERE ((dueDate IS NULL) > :DueIsNull) OR ((dueDate IS NULL)=:DueIsNull AND (dueDate, id) > (:DueDate, :ID))こっちが正しい。
				sb.WriteString("WHERE (due_is_null, dueDate, id) > (?, ?, ?)\n")
				dueIsNull := 0
				if q.Cursor.DueIsNull {
					dueIsNull = 1
				}
				args = append(args, dueIsNull, q.Cursor.DueDate, q.Cursor.ID.Value())

				//WHERE ((dueDate IS NOT NULL AND (dueDate > ? OR (dueDate = ? AND id > ?))) OR (dueDate IS NULL))これは式を含んでるから遅くなる。columnにする。
				//args = append(args, q.Cursor.DueDate, q.Cursor.DueDate, q.Cursor.ID.Value())
			} else {
				sb.WriteString("WHERE (dueDate IS NULL AND id > ?)\n")
				args = append(args, q.Cursor.ID.Value())
			}

		default:
			return nil, nil, dtask.ErrInvalidSort
		}
	}

	switch q.Sort {
	case dtask.SortCreated:
		sb.WriteString("ORDER BY created DESC, id DESC\n")
	case dtask.SortDueDate:

		sb.WriteString("ORDER BY due_is_null ASC, dueDate ASC, id ASC\n")
		//ORDER BY (dueDate IS NULL) ASC, dueDate ASC, id ASC
	default:
		return nil, nil, dtask.ErrInvalidSort
	}
	//N+1
	dbLimit := q.Limit + 1
	sb.WriteString("LIMIT ?\n")
	args = append(args, dbLimit)

	sql := sb.String()
	//var records []*TaskRecord 学習用メモ　これの場合毎回新しいヒープが発生、巨大レコード、加工や共有したい場合
	var records []TaskRecord
	if err := r.q.SelectContext(ctx, &records, sql, args...); err != nil {
		return nil, nil, fmt.Errorf("taskrepo list select: %w", err)
	}
	hasNext := len(records) > q.Limit
	//いらん部分削る
	if hasNext {
		records = records[:q.Limit]
	}

	tasks := make([]*dtask.Task, 0, len(records))
	for i := range records {
		rec := &records[i]
		t, err := RecordToEntity(rec)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid task record id=%s: %w", rec.ID, err)
		}
		tasks = append(tasks, t)
	}
	if !hasNext {
		return tasks, nil, nil
	}
	next := buildNextCursor(q.Sort, tasks)
	return tasks, next, nil
}

func buildNextCursor(sort dtask.ListSort, tasks []*dtask.Task) *dtask.ListCursor {
	if len(tasks) == 0 {
		return nil
	}
	last := tasks[len(tasks)-1]

	c := &dtask.ListCursor{ID: last.ID()}

	switch sort {
	case dtask.SortCreated:
		c.Created = last.CreatedAt()
	case dtask.SortDueDate:
		dd := last.DueDate()
		c.DueIsNull = false

		c.DueDate = dd.Value()

	}
	return c
}

/*package taskrepo

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
	ファイルに切り出しても//sqlが増えると別いい
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
*/

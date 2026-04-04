package taskrepo

import (
	"context"
	"fmt"
	"strings"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func (r *repository) List(ctx context.Context, q dtask.ListQuery) ([]*dtask.Task, *dtask.ListCursor, error) {

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

	args := make([]any, 0, 6)

	sb.WriteString(`
SELECT
  id, user_id, title, description, status, due_date,
  created_at, updated_at, version
FROM task
WHERE user_id = ?
`)
	args = append(args, q.UserID.Value())

	if q.Cursor != nil {
		switch q.Sort {
		case dtask.SortCreated:
			sb.WriteString("AND (created_at, id) < (?, ?)\n")

			args = append(args, q.Cursor.Created, q.Cursor.ID.Value())

		case dtask.SortDueDate:
			if !q.Cursor.DueIsNull {

				sb.WriteString("AND (due_is_null, due_date, id) > (?, ?, ?)\n")
				dueIsNull := 0
				if q.Cursor.DueIsNull {
					dueIsNull = 1
				}
				args = append(args, dueIsNull, q.Cursor.DueDate, q.Cursor.ID.Value())

			} else {
				sb.WriteString("AND (due_date IS NULL AND id > ?)\n")
				args = append(args, q.Cursor.ID.Value())
			}

		default:
			return nil, nil, dtask.ErrInvalidSort
		}
	}

	switch q.Sort {
	case dtask.SortCreated:
		sb.WriteString("ORDER BY created_at DESC, id DESC\n")
	case dtask.SortDueDate:

		sb.WriteString("ORDER BY due_is_null ASC, due_date ASC, id ASC\n")

	default:
		return nil, nil, dtask.ErrInvalidSort
	}
	//N+1
	dbLimit := q.Limit + 1
	sb.WriteString("LIMIT ?\n")
	args = append(args, dbLimit)

	sql := sb.String()

	var records []TaskRecord
	if err := r.q.SelectContext(ctx, &records, sql, args...); err != nil {
		return nil, nil, fmt.Errorf("taskrepo list select: %w", err)
	}
	hasNext := len(records) > q.Limit

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

package taskrepo

import (
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type repository struct {
	q db.QueryerExecer
}

var _ dtask.TaskRepository = (*repository)(nil)

func New(db db.QueryerExecer) *repository {
	return &repository{
		q: db,
	}
}

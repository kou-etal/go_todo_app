package taskrepo

import (
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.QueryerExecer
} //packageのstructは同一。これやとsaveのexecuer不可

var _ dtask.TaskRepository = (*Repository)(nil)

func NewRepository(db db.QueryerExecer) *Repository {
	return &Repository{
		q: db,
	}
}

package taskeventrepo

import (
	dtaskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type repository struct {
	q db.QueryerExecer
}

var _ dtaskevent.TaskEventRepository = (*repository)(nil)

func New(db db.QueryerExecer) *repository {
	return &repository{
		q: db,
	}
}

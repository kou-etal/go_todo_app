package userrepo

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.QueryerExecer
}

//DB直接じゃなくてQueryer/Execerで抽象する

var _ duser.UserRepository = (*Repository)(nil)

func NewRepository(q db.QueryerExecer) *Repository {
	return &Repository{
		q: q,
	}
}

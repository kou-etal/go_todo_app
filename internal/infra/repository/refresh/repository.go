package refreshrepo

import (
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.QueryerExecer
}

var _ drefresh.RefreshTokenRepository = (*Repository)(nil)

func NewRepository(q db.QueryerExecer) *Repository {
	return &Repository{q: q}
}

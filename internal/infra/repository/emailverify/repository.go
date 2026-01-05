package emailverifyrepo

import (
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.QueryerExecer
}

var _ dverify.EmailVerificationRepository = (*Repository)(nil)

func NewRepository(q db.QueryerExecer) *Repository {
	return &Repository{
		q: q,
	}
}

package taskrepo

import (
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

// db.QueryerExecerをrepoメソッドの引数からstructのフィールドにすることでusecaseがdbをimportしなくなった
// これ良く使える方式
type Repository struct {
	q db.QueryerExecer
} //packageのstructは同一。これやとsaveのexecuer不可
//これq db.QueryerExecerを引数にすることで*sqlx.DB と *sqlx.Txの差し替えが可能
//NewRepository(*sqlx.DB)とトランザクションのNewRepository(*sqlx.Tx)

var _ dtask.TaskRepository = (*Repository)(nil)

func NewRepository(db db.QueryerExecer) *Repository {
	return &Repository{
		q: db,
	}
}

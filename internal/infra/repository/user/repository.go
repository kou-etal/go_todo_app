package userrepo

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
)

type Repository struct {
	q db.QueryerExecer
}

//DB直接じゃなくてQueryer/Execerで抽象するメリット
//トランザクションを扱いやすい。*sqlx.DBと*sqlx.Txを切り替え可能。
//DBを作らずに簡単にSelectContext/GetContext/ExecContext/NamedExecContext実装したstubを用いてテストできる。
//sqlx.dbはできること多いが抽象によってクエリ実行だけに制限し、責務を閉じる

var _ duser.UserRepository = (*Repository)(nil)

func NewRepository(q db.QueryerExecer) *Repository {
	return &Repository{
		q: q,
	}
}

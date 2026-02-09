package userrepo

import (
	"context"
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user" //ドメイン系はdで受け取る慣習
)

// ポインタと値　ポインタ->状態持つ場合、nil表現、コピー避ける、->entity、repository
// 値->状態持たない、軽い、nilありえない、ただの入れ物->record
// レシーバーは全部ポインタなわけない
func (r *Repository) Store(ctx context.Context, u *duser.User) error { //まず引数と返り値を考える
	const q = `
INSERT INTO user (
  id, email, password_hash, user_name, email_verified_at,
  created_at, updated_at, version
) VALUES (
  :id, :email, :password_hash, :user_name, :email_verified_at,
  :created_at, :updated_at, :version
);
`
	//これはポインタじゃなくて値返す
	rec := toRecord(u)

	_, err := r.q.NamedExecContext(ctx, q, rec) //これはrecordの右側に対応
	if err != nil {
		return fmt.Errorf("userrepo store execute: %w", err)
	}
	return nil
}

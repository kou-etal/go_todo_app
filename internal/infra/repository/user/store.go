package userrepo

import (
	"context"
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func (r *Repository) Store(ctx context.Context, u *duser.User) error {
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

package userrepo

import (
	"context"
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func (r *Repository) FindByEmail(ctx context.Context, email duser.UserEmail) (*duser.User, error) {
	const q = `
SELECT id, email, password_hash, user_name, email_verified_at,
       created_at, updated_at, version
FROM users
WHERE email = ?
LIMIT 1;
`
	var rec UserRecord
	if err := r.q.GetContext(ctx, &rec, q, email.Value()); err != nil {
		if isNotFound(err) {
			return nil, duser.ErrNotFound
		}
		return nil, fmt.Errorf("userrepo find by email: %w", err)
	}
	return toEntity(rec)
}

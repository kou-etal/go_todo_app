package refreshrepo

import (
	"context"
	"fmt"

	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
)

func (r *Repository) FindByTokenHashForUpdate(ctx context.Context, hash drefresh.TokenHash) (*drefresh.RefreshToken, error) {
	const q = `
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = ?
FOR UPDATE;
`
	var rec RefreshTokenRecord
	if err := r.q.GetContext(ctx, &rec, q, hash.Hash()); err != nil {
		if isNotFound(err) {
			return nil, drefresh.ErrNotFound
		}
		return nil, fmt.Errorf("refreshrepo find by token_hash for update: %w", err)
	}
	t, err := toEntity(rec)
	if err != nil {
		return nil, fmt.Errorf("refreshrepo map record to entity: %w", err)
	}
	return t, nil
}

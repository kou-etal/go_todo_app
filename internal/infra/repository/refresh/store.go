package refreshrepo

import (
	"context"
	"fmt"

	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
)

func (r *Repository) Store(ctx context.Context, t *drefresh.RefreshToken) error {
	const q = `
INSERT INTO refresh_tokens (
  id, user_id, token_hash, expires_at, revoked_at, created_at
) VALUES (
  :id, :user_id, :token_hash, :expires_at, :revoked_at, :created_at
);
`
	rec := toRecord(t)
	_, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("refreshrepo store execute: %w", err)
	}
	return nil
}

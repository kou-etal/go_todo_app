package refreshrepo

import (
	"context"
	"fmt"

	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
)

func (r *Repository) Update(ctx context.Context, t *drefresh.RefreshToken) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = :revoked_at
WHERE id = :id;
`
	rec := toRecord(t)
	_, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("refreshrepo update execute: %w", err)
	}
	return nil
}

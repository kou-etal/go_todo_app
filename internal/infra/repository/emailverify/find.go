package emailverifyrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

func (r *Repository) FindByTokenHashForUpdate(ctx context.Context, hash dverify.TokenHash) (*dverify.EmailVerificationToken, error) {

	const q = `
SELECT
  id,
  user_id,
  token_hash,
  expires_at,
  used_at,
  created_at
FROM email_verification_tokens
WHERE token_hash = ?
FOR UPDATE;
`

	var rec EmailVerificationTokenRecord
	if err := r.q.GetContext(ctx, &rec, q, hash.Hash()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dverify.ErrNotFound
		}
		return nil, fmt.Errorf("emailverifyrepo find by token_hash for update: %w", err)
	}

	t, err := toEntity(rec)

	if err != nil {
		return nil, fmt.Errorf("emailverifyrepo map record to entity: %w", err)
	}

	return t, nil
}

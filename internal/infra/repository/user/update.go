package userrepo

import (
	"context"
	"fmt"
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func (r *Repository) MarkEmailVerified(
	ctx context.Context,
	userID duser.UserID,
	verifiedAt time.Time,
) error {

	const q = `
UPDATE user
SET
  email_verified_at = ?,
  updated_at = ?,
  version = version + 1
WHERE
  id = ?
  AND email_verified_at IS NULL;
`

	res, err := r.q.ExecContext(
		ctx,
		q,
		verifiedAt,
		verifiedAt,
		userID.Value(),
	)
	if err != nil {
		return fmt.Errorf("userrepo mark email verified execute: %w", err)
	}

	ra, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("userrepo mark email verified rowsaffected: %w", err)
	}

	if ra == 0 {

		return duser.ErrConflict
	}

	return nil
}

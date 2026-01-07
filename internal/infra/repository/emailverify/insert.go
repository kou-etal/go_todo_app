package emailverifyrepo

//insert 新規作成だけで更新しない意味合い。
//store　更新も含む意味合い。
//今回はinsert
import (
	"context"
	"fmt"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

func (r *Repository) Insert(ctx context.Context, t *dverify.EmailVerificationToken) error {

	const q = `
INSERT INTO email_verification_tokens (
  id,
  user_id,
  token_hash,
  expires_at,
  used_at,
  created_at
) VALUES (
  :id,
  :user_id,
  :token_hash,
  :expires_at,
  :used_at,
  :created_at
);
`

	rec := toRecord(t)

	_, err := r.q.NamedExecContext(ctx, q, rec)
	if err != nil {
		return fmt.Errorf("emailverifyrepo insert execute: %w", err)
	}
	return nil
}

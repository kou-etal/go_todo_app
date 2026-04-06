package emailverifyrepo //命名怪しい。

import (
	"time"
)

type EmailVerificationTokenRecord struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	TokenHash string     `db:"token_hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"` //TODO:sql.NullTimeが適切
	CreatedAt time.Time  `db:"created_at"`
}

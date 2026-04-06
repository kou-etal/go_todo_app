package refresh

import (
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

type RefreshToken struct {
	id        TokenID
	userID    duser.UserID
	tokenHash TokenHash
	expiresAt time.Time
	revokedAt *time.Time
	createdAt time.Time
}

func (t *RefreshToken) ID() TokenID        { return t.id }
func (t *RefreshToken) UserID() duser.UserID { return t.userID }
func (t *RefreshToken) TokenHash() TokenHash { return t.tokenHash }
func (t *RefreshToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *RefreshToken) RevokedAt() *time.Time { return t.revokedAt }
func (t *RefreshToken) CreatedAt() time.Time { return t.createdAt }

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(t.expiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
	return t.revokedAt != nil
}

func (t *RefreshToken) IsValid(now time.Time) bool {
	return !t.IsExpired(now) && !t.IsRevoked()
}

func (t *RefreshToken) Revoke(now time.Time) {
	n := normalizeTime(now)
	t.revokedAt = &n
}

func normalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

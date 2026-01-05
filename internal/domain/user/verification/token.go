// password reset,session,refresh token,MFA,OAuthとか作るならdomain/でトップレベルドメイン切ってもいい。
package verification

import (
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

type EmailVerificationToken struct {
	id        TokenID
	userID    duser.UserID
	tokenHash TokenHash
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
}

func (t *EmailVerificationToken) ID() TokenID          { return t.id }
func (t *EmailVerificationToken) UserID() duser.UserID { return t.userID }
func (t *EmailVerificationToken) TokenHash() TokenHash { return t.tokenHash }
func (t *EmailVerificationToken) ExpiresAt() time.Time { return t.expiresAt }
func (t *EmailVerificationToken) UsedAt() *time.Time   { return t.usedAt }
func (t *EmailVerificationToken) CreatedAt() time.Time { return t.createdAt }

// これは相対の状態遷移ロジック、factoryでもvoでもない。entityに置く。
func (t *EmailVerificationToken) IsExpired(now time.Time) bool {
	return !now.Before(t.expiresAt)
}

func (t *EmailVerificationToken) IsUsed() bool {
	return t.usedAt != nil
}

func (t *EmailVerificationToken) IsValid(now time.Time) bool {
	return !t.IsExpired(now) && !t.IsUsed()
}

func (t *EmailVerificationToken) Consume(now time.Time) {
	n := normalizeTime(now)
	t.usedAt = &n
}
func normalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

//tokenはライフサイクル短い、更新もされないからchangeメソッド作らない。

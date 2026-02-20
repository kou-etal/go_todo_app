package verification

import (
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

const defaultTokenTTL = 24 * time.Hour

func NewEmailVerificationToken(
	userID duser.UserID,
	hasher TokenHasher,
	gen TokenGenerator,
	now time.Time,
) (*EmailVerificationToken, string, error) {

	n := normalizeTime(now)

	plain, err := gen.Generate()

	if err != nil {
		return nil, "", err
	}
	hash, err := NewUserTokenHashFromPlain(plain, hasher)
	if err != nil {
		return nil, "", err
	}
	return &EmailVerificationToken{
		id:        NewTokenID(),
		userID:    userID,
		tokenHash: hash,
		expiresAt: n.Add(defaultTokenTTL),
		usedAt:    nil,
		createdAt: n,
	}, plain, nil
}

// これは復元用。repoで使う
// domainはDBが持ってる値を信用する
func ReconstructEmailVerificationToken(
	id TokenID,
	userID duser.UserID,
	tokenHash TokenHash,
	expiresAt time.Time,
	usedAt *time.Time,
	createdAt time.Time,
) *EmailVerificationToken {
	return &EmailVerificationToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		usedAt:    usedAt,
		createdAt: createdAt,
	}
}

package refresh

import (
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func NewRefreshToken(
	userID duser.UserID,
	hasher TokenHasher,
	gen TokenGenerator,
	ttl time.Duration,
	now time.Time,
) (*RefreshToken, string, error) {
	n := normalizeTime(now)

	plain, err := gen.Generate()
	if err != nil {
		return nil, "", err
	}
	hash, err := NewTokenHashFromPlain(plain, hasher)
	if err != nil {
		return nil, "", err
	}
	return &RefreshToken{
		id:        NewTokenID(),
		userID:    userID,
		tokenHash: hash,
		expiresAt: n.Add(ttl),
		revokedAt: nil,
		createdAt: n,
	}, plain, nil
}

func ReconstructRefreshToken(
	id TokenID,
	userID duser.UserID,
	tokenHash TokenHash,
	expiresAt time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) *RefreshToken {
	return &RefreshToken{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		expiresAt: expiresAt,
		revokedAt: revokedAt,
		createdAt: createdAt,
	}
}

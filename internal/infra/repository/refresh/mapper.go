package refreshrepo

import (
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	drefresh "github.com/kou-etal/go_todo_app/internal/domain/user/refresh"
)

func toEntity(r RefreshTokenRecord) (*drefresh.RefreshToken, error) {
	id, err := drefresh.ParseTokenID(r.ID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid refresh_token record id=%s field=id: %w",
			r.ID, err,
		)
	}
	userID, err := duser.ParseUserID(r.UserID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid refresh_token record id=%s field=user_id: %w",
			r.ID, err,
		)
	}
	tokenHash, err := drefresh.ReconstructTokenHash(r.TokenHash)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid refresh_token record id=%s field=token_hash: %w",
			r.ID, err,
		)
	}
	return drefresh.ReconstructRefreshToken(
		id,
		userID,
		tokenHash,
		r.ExpiresAt,
		r.RevokedAt,
		r.CreatedAt,
	), nil
}

func toRecord(t *drefresh.RefreshToken) RefreshTokenRecord {
	return RefreshTokenRecord{
		ID:        t.ID().Value(),
		UserID:    t.UserID().Value(),
		TokenHash: t.TokenHash().Hash(),
		ExpiresAt: t.ExpiresAt(),
		RevokedAt: t.RevokedAt(),
		CreatedAt: t.CreatedAt(),
	}
}

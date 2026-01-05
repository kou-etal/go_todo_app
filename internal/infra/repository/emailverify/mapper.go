package emailverifyrepo

import (
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

func toEntity(r EmailVerificationTokenRecord) (*dverify.EmailVerificationToken, error) {

	id, err := dverify.ParseTokenID(r.ID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid email_verification_token record id=%s field=id: %w",
			r.ID, err,
		)
	}
	//これdomain横断するのどうなん。verificationにもparseuserid作ってそれ使う。そこまでしなあかんか。
	userID, err := duser.ParseUserID(r.UserID)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid email_verification_token record id=%s field=user_id: %w",
			r.ID, err,
		)
	}
	//token_hashはログに出さない
	tokenHash, err := dverify.ReconstructTokenHash(r.TokenHash)
	if err != nil {
		return nil, fmt.Errorf(
			"invalid email_verification_token record id=%s field=token_hash: %w",
			r.ID, err,
		)
	}

	return dverify.ReconstructEmailVerificationToken(
		id,
		userID,
		tokenHash,
		r.ExpiresAt,
		r.UsedAt,
		r.CreatedAt,
	), nil
}
func toRecord(t *dverify.EmailVerificationToken) EmailVerificationTokenRecord {

	return EmailVerificationTokenRecord{
		ID:        t.ID().Value(),
		UserID:    t.UserID().Value(),
		TokenHash: t.TokenHash().Hash(),
		ExpiresAt: t.ExpiresAt(),
		UsedAt:    t.UsedAt(),
		CreatedAt: t.CreatedAt(),
	}
}

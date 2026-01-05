package verification

import (
	"time"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

const defaultTokenTTL = 24 * time.Hour

//今までfactoryはerror返さなかったがそれはusecaseで保証しきった値を代入してた
//今回はgenerateはdomain責務。usecaseでは作れない。errorを返す。
//それとplainも欲しい。hash voはhashを生成する。そこにplain返す責務持たせるの良くない。factoryで返す。

func NewEmailVerificationToken(
	userID duser.UserID,
	hash TokenHash,
	gen TokenGenerator,
	now time.Time,
) (*EmailVerificationToken, string, error) {

	n := normalizeTime(now)

	plain, err := gen.Generate()
	//plain は RandomTokenGeneratorが保証。バリデーションしない。
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

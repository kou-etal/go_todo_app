package user

import (
	"errors"
	"strings"
)

type UserPassword struct {
	hash string
}

// hashの抽象はドメイン責務、実装はinfra層にしてる。usecaseに抽象置くことも多い。
// これhasherの抽象。password_hasher.goにおいても良い。passwordが厚くなったら。
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// 変換+normalization
// domainではplainではなくhashを扱う。
func NewUserPasswordFromPlain(
	plain string,
	hasher PasswordHasher,
) (UserPassword, error) {

	if strings.TrimSpace(plain) == "" {
		return UserPassword{}, ErrEmptyPassword
	}
	//passwordではスペースは丸めるではなく、そもそも弾く
	//descriptionとかはスペース丸めてる。
	if plain != strings.TrimSpace(plain) {
		return UserPassword{}, ErrPasswordHasLeadingOrTrailingSpace
	}
	//パスワード長はbyteで評価する(runecountではない)。bcyptがbyteを扱う。
	//runecountはUIの都合
	if len(plain) < 12 {
		return UserPassword{}, ErrPasswordTooShort
	}
	if len(plain) > 72 {
		//bcryptの制限
		return UserPassword{}, ErrPasswordTooLong
	}

	hash, err := hasher.Hash(plain)
	if err != nil {
		return UserPassword{}, err
	}

	return UserPassword{hash: hash}, nil
}

// これrepoから使う。passwordでは普通はusecase->domain repo->domainを再利用するところを分けて実装する。
func ReconstructUserPassword(hash string) (UserPassword, error) {
	if hash == "" {
		return UserPassword{}, errors.New("password hash is empty")
	}
	return UserPassword{hash: hash}, nil
}

func (p UserPassword) Hash() string {
	return p.hash
}

// plainだけやったらどのhasherかわからん。
func (p UserPassword) Compare(
	plain string,
	hasher PasswordHasher,
) error {
	return hasher.Compare(p.hash, plain)
}

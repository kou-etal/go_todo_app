package verification

import (
	"errors"
)

type TokenHash struct {
	hash string
}

// hashの抽象はドメイン責務、実装はinfra層にしてる。usecaseに抽象置くことも多い。
type TokenHasher interface {
	Hash(plain string) (string, error)
}

type TokenGenerator interface {
	Generate() (string, error)
} //永続の抽象ではないからここに置く

// 変換+normalization
// domainではplainではなくhashを扱う。
func NewUserTokenHashFromPlain(
	plain string,
	hasher TokenHasher,
) (TokenHash, error) {
	//ユーザー入力じゃないからgeneraterから受け取ったplainはvalidationしなくていい。
	if plain == "" {
		return TokenHash{}, ErrEmptyToken
	}
	hash, err := hasher.Hash(plain)
	if err != nil {
		return TokenHash{}, err
	}

	return TokenHash{hash: hash}, nil
}

// これrepoから使う。普通はusecase->domain repo->domainを再利用するところを分けて実装する。
func ReconstructTokenHash(hash string) (TokenHash, error) {
	if hash == "" {
		return TokenHash{}, errors.New("token hash is empty")
	}
	return TokenHash{hash: hash}, nil
}

func (t TokenHash) Hash() string {
	return t.hash
}

//compareは不要
/*認証フロー
まずユーザー作成+リンクにリンクにplain token載せてメール送る。userテーブル emailverifiedat=null tokenテーブル usedat=null トランザクション
次にメールからリンク踏む　ここでplaintokenをhashにする。hashをdbから検索(ここcompareではない)。
userテーブル emailverifiedat設定　トランザクション
*/

package security

//外部ライブラリに近いからinfraに置く。
import (
	user "github.com/kou-etal/go_todo_app/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

//bcryptはパスワード専用の暗号ハッシュ、SHA256のような一般ハッシュとは違う

type BcryptHasher struct {
	cost int
	//これはどれだけ遅くするかの設定。遅ければ遅いほど解読されにくいが重い。10-14ぐらいが良い
}

var _ user.PasswordHasher = (*BcryptHasher)(nil)

// costを受け取る。
func NewBcryptHasher(cost int) *BcryptHasher {
	//これcostバリデーションしたほうがいい。NewBcryptHasher(-1)が通る。
	return &BcryptHasher{cost: cost}
}

// plain->hash
// 適当にポインタレシーバにしがち
func (h *BcryptHasher) Hash(plain string) (string, error) {
	//[]byte(plain) バイト文字列で受け取る。
	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)
	//これ返り値ASCII文字列やからstring可能
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//DBに保存しやすいようにstringにして返す

// ログイン用の比較メソッド
func (h *BcryptHasher) Compare(hash, plain string) error {
	//これなんでerr返さない。
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	//わざわざtrueとかで返さないのがGolang
}

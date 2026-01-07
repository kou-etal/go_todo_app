package security

//外部ライブラリに近いからinfraに置く。
import (
	"crypto/sha256" //これ標準ライブラリ
	"encoding/hex"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification" //命名これで良い
)

// tokenはSHA256使う。DB検索可能。
type SHA256TokenHasher struct{}

// state持たないからnew作らない
var _ dverify.TokenHasher = (*SHA256TokenHasher)(nil)

// plain->hash
// 適当にポインタレシーバにしがち
// state を持たないから値レシーバ。コピーのコストもない
func (h SHA256TokenHasher) Hash(plain string) (string, error) {
	sum := sha256.Sum256([]byte(plain))
	//これsumはASCII文字列ではない。バイト文字列[32]byte。各バイトを16進数2文字で表現してからstringこれがhex。
	// 引数はバイトではなくスライスやからsum[:]
	return hex.EncodeToString(sum[:]), nil
}

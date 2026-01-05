package security

//hex->1バイト2文字　安全重視 token
//base64->3バイト4文字　効率重視

import (
	"crypto/rand"
	"encoding/base64"
	"io"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

type RandomTokenGenerator struct {
	byteLen int
}

var _ dverify.TokenGenerator = (*RandomTokenGenerator)(nil)

// 32バイトぐらいが多い。
func NewRandomTokenGenerator(byteLen int) *RandomTokenGenerator {
	return &RandomTokenGenerator{
		byteLen: byteLen,
	}
}

func (g *RandomTokenGenerator) Generate() (string, error) {
	b := make([]byte, g.byteLen)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}

	//URL安全。base64
	return base64.RawURLEncoding.EncodeToString(b), nil
}

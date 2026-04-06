package security

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

	return base64.RawURLEncoding.EncodeToString(b), nil
}

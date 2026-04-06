package security

import (
	"crypto/sha256"
	"encoding/hex"

	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

type SHA256TokenHasher struct{}

var _ dverify.TokenHasher = (*SHA256TokenHasher)(nil)

func (h SHA256TokenHasher) Hash(plain string) (string, error) {
	sum := sha256.Sum256([]byte(plain))

	return hex.EncodeToString(sum[:]), nil
}

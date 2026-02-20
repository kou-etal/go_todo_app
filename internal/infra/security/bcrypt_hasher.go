package security

import (
	user "github.com/kou-etal/go_todo_app/internal/domain/user"
	"golang.org/x/crypto/bcrypt"
)

type BcryptHasher struct {
	cost int
}

var _ user.PasswordHasher = (*BcryptHasher)(nil)

func NewBcryptHasher(cost int) *BcryptHasher {

	return &BcryptHasher{cost: cost}
}

func (h *BcryptHasher) Hash(plain string) (string, error) {

	b, err := bcrypt.GenerateFromPassword([]byte(plain), h.cost)

	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *BcryptHasher) Compare(hash, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))

}

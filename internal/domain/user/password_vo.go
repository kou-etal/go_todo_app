package user

import (
	"errors"
	"strings"
)

type UserPassword struct {
	hash string
}

type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

func NewUserPasswordFromPlain(
	plain string,
	hasher PasswordHasher,
) (UserPassword, error) {

	if strings.TrimSpace(plain) == "" {
		return UserPassword{}, ErrEmptyPassword
	}

	if plain != strings.TrimSpace(plain) {
		return UserPassword{}, ErrPasswordHasLeadingOrTrailingSpace
	}

	if len(plain) < 12 {
		return UserPassword{}, ErrPasswordTooShort
	}
	if len(plain) > 72 {

		return UserPassword{}, ErrPasswordTooLong
	}

	hash, err := hasher.Hash(plain)
	if err != nil {
		return UserPassword{}, err
	}

	return UserPassword{hash: hash}, nil
}

func ReconstructUserPassword(hash string) (UserPassword, error) {
	if hash == "" {
		return UserPassword{}, errors.New("password hash is empty")
	}
	return UserPassword{hash: hash}, nil
}

func (p UserPassword) Hash() string {
	return p.hash
}

func (p UserPassword) Compare(
	plain string,
	hasher PasswordHasher,
) error {
	return hasher.Compare(p.hash, plain)
}

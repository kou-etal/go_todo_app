package verification

import (
	"errors"
)

type TokenHash struct {
	hash string
}

type TokenHasher interface {
	Hash(plain string) (string, error)
}

type TokenGenerator interface {
	Generate() (string, error)
}

func NewUserTokenHashFromPlain(
	plain string,
	hasher TokenHasher,
) (TokenHash, error) {

	if plain == "" {
		return TokenHash{}, ErrEmptyToken
	}
	hash, err := hasher.Hash(plain)
	if err != nil {
		return TokenHash{}, err
	}

	return TokenHash{hash: hash}, nil
}

func ReconstructTokenHash(hash string) (TokenHash, error) {
	if hash == "" {
		return TokenHash{}, errors.New("token hash is empty")
	}
	return TokenHash{hash: hash}, nil
}

func (t TokenHash) Hash() string {
	return t.hash
}

package verification

import (
	"github.com/google/uuid"
)

type TokenID string

func NewTokenID() TokenID {
	return TokenID(uuid.New().String())
}

func ParseTokenID(v string) (TokenID, error) {
	_, err := uuid.Parse(v)
	if err != nil {
		return "", ErrInvalidTokenID
	}
	return TokenID(v), nil
}

func (id TokenID) Value() string {
	return string(id)
}

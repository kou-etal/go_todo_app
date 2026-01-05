package verification

import "errors"

var (
	ErrInvalidTokenID = errors.New("invalid ID")
	ErrEmptyToken     = errors.New("empty token")
)

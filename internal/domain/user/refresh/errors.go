package refresh

import "errors"

var (
	ErrInvalidTokenID = errors.New("invalid token ID")
	ErrEmptyToken     = errors.New("empty token")
	ErrNotFound       = errors.New("refresh token not found")
)

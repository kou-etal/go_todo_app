package verification

import "errors"

var (
	ErrInvalidTokenID = errors.New("invalid ID")
	ErrEmptyToken     = errors.New("empty token")
	ErrInvalidID      = errors.New("invalid ID")
	ErrNotFound       = errors.New("not found")
	ErrConflict       = errors.New("taskrepo: conflict")
)

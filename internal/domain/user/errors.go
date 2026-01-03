package user

import "errors"

var (
	ErrEmptyEmail                        = errors.New("email is empty")
	ErrEmailTooLong                      = errors.New("email is too long")
	ErrInvalidEmailFormat                = errors.New("invalid email format")
	ErrInvalidID                         = errors.New("invalid ID")
	ErrEmptyPassword                     = errors.New("password is empty")
	ErrPasswordTooShort                  = errors.New("password is too short")
	ErrPasswordTooLong                   = errors.New("password is too long")
	ErrPasswordHasLeadingOrTrailingSpace = errors.New("password has leading or trailing whitespace")
	ErrEmptyName                         = errors.New("name is empty")
	ErrNameTooLong                       = errors.New("name is too long")
)

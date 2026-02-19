package register

import (
	"errors"
	"strings"
)

var (
	ErrEmptyEmail                        = errors.New("email is required")
	ErrEmptyPassword                     = errors.New("password is required")
	ErrEmptyUserName                     = errors.New("user name is required")
	ErrEmailTooLong                      = errors.New("email too long")
	ErrInvalidEmailFormat                = errors.New("invalid email format")
	ErrPasswordTooShort                  = errors.New("password too short")
	ErrPasswordTooLong                   = errors.New("password too long")
	ErrPasswordHasLeadingOrTrailingSpace = errors.New("password has leading or trailing space")
	ErrNameTooLong                       = errors.New("name too long")
	ErrConflict                          = errors.New("conflict")
)
func normalize(cmd Command) (Command, error) {
	cmd.Email = strings.TrimSpace(cmd.Email)
	cmd.UserName = strings.TrimSpace(cmd.UserName)

	//passwordはtrimしない
	//normalizeとvalidateの違い
	//passwordはvalidateだけ。

	if cmd.Email == "" {
		return Command{}, ErrEmptyEmail
	}
	if cmd.Password == "" {
		return Command{}, ErrEmptyPassword
	}
	if cmd.UserName == "" {
		return Command{}, ErrEmptyUserName
	}

	return cmd, nil
}

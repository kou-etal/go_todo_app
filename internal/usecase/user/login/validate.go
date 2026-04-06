package login

import (
	"errors"
	"strings"
)

var (
	ErrEmptyEmail          = errors.New("email is required")
	ErrEmptyPassword       = errors.New("password is required")
	ErrInvalidCredentials  = errors.New("invalid credentials")
)

func normalize(cmd Command) (Command, error) {
	cmd.Email = strings.TrimSpace(cmd.Email)

	if cmd.Email == "" {
		return Command{}, ErrEmptyEmail
	}
	if cmd.Password == "" {
		return Command{}, ErrEmptyPassword
	}

	return cmd, nil
}

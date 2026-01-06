package register

import (
	"errors"
	"strings"
)

var (
	ErrEmptyEmail    = errors.New("email is required")
	ErrEmptyPassword = errors.New("password is required")
	ErrEmptyUserName = errors.New("user name is required")
)

// TODO:巨大データは弾いてもいい。taskではそうした。
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

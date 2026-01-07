package user

import (
	"strings"
	"unicode/utf8"
)

type UserName struct {
	value string
}

func NewUserName(v string) (UserName, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return UserName{}, ErrEmptyName
	}

	const maxNameLength = 20
	if utf8.RuneCountInString(v) > maxNameLength {
		return UserName{}, ErrNameTooLong
	}
	return UserName{value: v}, nil
}

func (n UserName) Value() string {
	return n.value
}

package user

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

type UserEmail struct {
	value string
}

func NewUserEmail(v string) (UserEmail, error) {
	v = strings.TrimSpace(v)
	if v == "" {
		return UserEmail{}, ErrEmptyEmail
	}

	v = strings.ToLower(v)
	const maxEmailLength = 254
	if utf8.RuneCountInString(v) > maxEmailLength {
		return UserEmail{}, ErrEmailTooLong
	}
	if !isValidEmailFormat(v) {
		return UserEmail{}, ErrInvalidEmailFormat
	}
	return UserEmail{value: v}, nil
}

func (t UserEmail) Value() string {
	return t.value
}
func isValidEmailFormat(v string) bool {
	_, err := mail.ParseAddress(v)
	return err == nil
}

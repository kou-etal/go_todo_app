package user

import (
	"github.com/google/uuid"
)

type UserID string

func NewUserID() UserID {
	return UserID(uuid.New().String())
}

// string->userid変換、生成ではない。これはstringやからここにおける。
func ParseUserID(v string) (UserID, error) {
	_, err := uuid.Parse(v)
	if err != nil {
		return "", ErrInvalidID
	}
	return UserID(v), nil
}
func (id UserID) Value() string {
	return string(id)
}

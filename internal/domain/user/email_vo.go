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

	//大文字のnormalization
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

//もしupdateを部分送信ではなく全部送信の仕様ならばequalsはあってもいい。
// 更新とかの時にもし変更なしなら更新しないってのが簡単に記述できる。あとテストの時にも使える。
/*func (e UserEmail) Equals(other UserEmail) bool {
    return e.value == other.value
}*/
/*if task.Title().Equals(newTitle) {
    return nil
}
task.ChangeTitle(newTitle, now)
*/

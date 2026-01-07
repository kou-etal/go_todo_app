package user

import (
	"time"
)

func NewUser(
	email UserEmail,
	password UserPassword,
	userName UserName,
	now time.Time,
) *User {
	n := normalizeTime(now)

	return &User{
		id:              NewUserID(), //IDが欲しいのはdomain層の都合
		email:           email,
		passwordHash:    password,
		userName:        userName,
		emailVerifiedAt: nil, //ここを変えれるのはverifyemailメソッドだけ。これを適当に変えさせないためにカプセル
		createdAt:       n,
		updatedAt:       n,
		version:         1,
	}
}

// これは復元用。repoで使う
// domainはDBが持ってる値を信用する
func ReconstructUser(
	id UserID,
	email UserEmail,
	password UserPassword,
	userName UserName,
	emailVerifiedAt *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	version uint64,
) *User {
	return &User{
		id:              id,
		email:           email,
		passwordHash:    password,
		emailVerifiedAt: emailVerifiedAt,
		userName:        userName,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		version:         version,
	}
}

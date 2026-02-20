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
		id:              NewUserID(),
		email:           email,
		passwordHash:    password,
		userName:        userName,
		emailVerifiedAt: nil,
		createdAt:       n,
		updatedAt:       n,
		version:         1,
	}
}

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

package user

import (
	"time"
)

type User struct {
	//すべて必須。そうでない場合userName *UserName
	id              UserID
	email           UserEmail
	password        UserPassword
	userName        UserName
	emailVerifiedAt *time.Time
	createdAt       time.Time
	updatedAt       time.Time
	version         uint64 //++、楽観ロックはDB責務
}

func (t *User) ID() UserID                  { return t.id }
func (t *User) Email() UserEmail            { return t.email }
func (t *User) Password() UserPassword      { return t.password }
func (t *User) UserName() UserName          { return t.userName }
func (t *User) EmailVerifiedAt() *time.Time { return t.emailVerifiedAt }
func (t *User) CreatedAt() time.Time        { return t.createdAt }
func (t *User) UpdatedAt() time.Time        { return t.updatedAt }
func (t *User) Version() uint64             { return t.version }

// email変更は重いから簡単に実行しない
func (t *User) ChangeEmail(newEmail UserEmail, now time.Time) {
	t.email = newEmail
	t.emailVerifiedAt = nil
	t.updateTime(now)
}
func (t *User) ChangePassword(newPass UserPassword, now time.Time) {
	t.password = newPass
	t.updateTime(now)
}
func (t *User) ChangeUserName(newName UserName, now time.Time) {
	t.userName = newName
	t.updateTime(now)
}
func (u *User) VerifyEmail(now time.Time) {
	n := normalizeTime(now)
	u.emailVerifiedAt = &n
	u.updatedAt = n
}
func (t *User) updateTime(now time.Time) {
	n := normalizeTime(now)
	t.updatedAt = n
}

func normalizeTime(t time.Time) time.Time {
	return t.UTC().Truncate(time.Second)
}

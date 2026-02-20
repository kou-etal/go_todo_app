package userrepo

import (
	"fmt"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func toEntity(r UserRecord) (*duser.User, error) { //duser.Userはコピーを作らせたくないからポインタ

	id, err := duser.ParseUserID(r.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=id: %w", r.ID, err)
	}
	//ここのnewはvalidation+型変換のうちただの型からdomainの型に変換する意味。usecaseから使うnewとは意味合い違うけど再利用。
	//TODO:でもよく考えたらもしvoに新たな制約つくったら既存のデータ復元できなくなる。ゆえに全部にreconstruct作るのが丸いかも

	email, err := duser.NewUserEmail(r.Email)
	if err != nil {

		return nil, fmt.Errorf("invalid user record id=%s field=email: %w", r.ID, err)
	}
	password, err := duser.ReconstructUserPassword(r.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=password: %w", r.ID, err)
	}
	name, err := duser.NewUserName(r.UserName)
	if err != nil {
		return nil, fmt.Errorf("invalid user record id=%s field=userName: %w", r.ID, err)
	}

	return duser.ReconstructUser(
		id,
		email,
		password,
		name,
		r.EmailVerifiedAt,
		r.Created,
		r.Updated,
		r.Version,
	), nil
}

func toRecord(u *duser.User) UserRecord {

	return UserRecord{
		ID:              u.ID().Value(),
		Email:           u.Email().Value(),
		PasswordHash:    u.Password().Hash(),
		UserName:        u.UserName().Value(),
		EmailVerifiedAt: u.EmailVerifiedAt(),
		Created:         u.CreatedAt(),
		Updated:         u.UpdatedAt(),
		Version:         u.Version(),
	}
}

package txdeps

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type registerdeps struct {
	userRepo        duser.UserRepository
	emailVerifyRepo dverify.EmailVerificationRepository
} //usecaseで操作されたくないから大文字にしない。

var _ usetx.RegisterDeps = (*registerdeps)(nil)

func NewRegister(
	u duser.UserRepository,
	v dverify.EmailVerificationRepository,
) usetx.RegisterDeps {
	return &registerdeps{
		userRepo:        u,
		emailVerifyRepo: v,
	}
}

func (d *registerdeps) UserRepo() duser.UserRepository {
	return d.userRepo
}
func (d *registerdeps) EmailVerifyRepo() dverify.EmailVerificationRepository {
	return d.emailVerifyRepo
}

package txdeps

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

//そもそもdepsの目的はトランザクションの中で使うメソッドを統合的に使える構造体を作ること。そのためにreturn d.userRepoしてる。
//じゃあdeps無しでrepo2個をfnに引数で渡せばいいのでは。->repoが増えるたびにハードコーディング
//tx増やしたくなったらnewファイル並べればいい

//deps具体は配線->appに置くがrunner具体はDBの関心->infraに置く

type registerdeps struct {
	userRepo        duser.UserRepository
	emailVerifyRepo dverify.EmailVerificationRepository
} //usecaseで操作されたくないから大文字にしない。usecaseに許可するのはinterface経由のgetterだけ。deps.UserRepo=を防ぐ
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

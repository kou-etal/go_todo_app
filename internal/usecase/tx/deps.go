package tx

import (
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

type RegisterDeps interface {
	UserRepo() duser.UserRepository
	EmailVerifyRepo() dverify.EmailVerificationRepository
} //トランザクション新しく作る場合は

/*
type RegisterDepsDeps struct {
	UserRepo        duser.UserRepository
	EmailVerifyRepo dverify.EmailVerificationRepository
}
*/

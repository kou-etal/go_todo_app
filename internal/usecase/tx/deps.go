package tx

import (
	dtaskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	dverify "github.com/kou-etal/go_todo_app/internal/domain/user/verification"
)

// deps抽象
type RegisterDeps interface {
	UserRepo() duser.UserRepository
	EmailVerifyRepo() dverify.EmailVerificationRepository
}
type TaskEventDeps interface {
	TaskRepo() dtask.TaskRepository
	TaskEventRepo() dtaskevent.TaskEventRepository
}

//トランザクション新しく作る場合はdepsを増やす

/*

type RegisterDepsDeps struct {
	UserRepo        duser.UserRepository
	EmailVerifyRepo dverify.EmailVerificationRepository
}
*/

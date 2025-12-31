package delete

import (
	"context"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

type Usecase struct {
	repo dtask.TaskRepository
}

func New(repo dtask.TaskRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Do(ctx context.Context, cmd Command) error {
	//deleteはresult返さずにステータスメッセージだけ
	//findbyidもしない。ただ消すだけ
	cmd, err := normalize(cmd)
	if err != nil {
		return err
	}

	id, err := dtask.ParseTaskID(cmd.ID)
	if err != nil {
		return err
	}

	//楽観ロックはrepoに寄せる。
	if err := u.repo.Delete(ctx, id, cmd.Version); err != nil {
		return err
	}

	return nil
}

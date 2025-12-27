package update

import (
	"context"

	"github.com/kou-etal/go_todo_app/internal/clock"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

type Usecase struct {
	repo  dtask.TaskRepository
	clock clock.Clocker
}

func New(repo dtask.TaskRepository, clock clock.Clocker) *Usecase {
	return &Usecase{repo: repo, clock: clock}
}
func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {

	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err

	}
	id, err := dtask.ParseTaskID(cmd.ID)
	if err != nil {
		return Result{}, err
	}
	t, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return Result{}, err
	}
	//sqlでも確認するがこっちでも調べる。楽観ロック
	if t.Version() != cmd.Version {
		//両方uint
		return Result{}, dtask.ErrConflict
	}
	now := u.clock.Now()

	if cmd.Title != nil {
		title, err := dtask.NewTaskTitle(*cmd.Title)
		if err != nil {
			return Result{}, err
		}
		t.ChangeTitle(title, now) //更新系ヘルパーはdomainで定義
	}
	if cmd.Description != nil {
		desc, err := dtask.NewTaskDescription(*cmd.Description)
		if err != nil {
			return Result{}, err
		}
		t.ChangeDescription(desc, now)
	}
	if cmd.DueDate != nil {
		opt, err := normalizeDueOption(*cmd.DueDate)
		if err != nil {
			return Result{}, err
		}
		due, err := dtask.NewDueDateFromOption(now, opt)
		if err != nil {
			return Result{}, err
		}
		t.ChangeDueDate(due, now)
	}
	if err := u.repo.Update(ctx, t); err != nil {
		return Result{}, err
	}
	return Result{
		ID: t.ID().Value(),
		//TODO:versionも返すべき。その場合repoに+の責務を寄せてるから更新後selectが必須。
	}, nil
}

func normalizeDueOption(t int) (dtask.DueOption, error) {
	switch t {
	case 7:
		return dtask.Due7Days, nil
	case 14:
		return dtask.Due14Days, nil
	case 21:
		return dtask.Due21Days, nil
	case 30:
		return dtask.Due30Days, nil
	default:
		return 0, ErrInvalidDueOption //これって0で返していいん
	}

}

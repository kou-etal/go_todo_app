package create

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

// mapperは使わない。newtaskを使い
func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err
	} //usecaseのエラー
	title, err := dtask.NewTaskTitle(cmd.Title)
	if err != nil {

		return Result{}, err
	} //ここからはdomainで撃ち落としたエラー
	desc, err := dtask.NewTaskDescription(cmd.Description)
	if err != nil {
		return Result{}, err
	}
	now := u.clock.Now()

	due, err := dtask.NewDueDateFromOption(now, cmd.DueDate)
	//usecaseがclockでnow取得する責務にしてるから一貫させるためにここでもnowを与える。
	//でもこれの場合usecaseがdomainに関与しすぎてるっていう考え方もある。
	//結局はdomainを変えたらusecaseでの与え方も変えなければならない設計が良くない。それが密に結合してるの意味。
	//これは別にdomain変えても被害なしやからいい。
	if err != nil {
		return Result{}, err
	}
	t := dtask.NewTask(title, desc, due, now)
	if err := u.repo.Store(ctx, t); err != nil {
		return Result{}, err
	}
	return Result{ID: t.ID().Value()}, nil
}

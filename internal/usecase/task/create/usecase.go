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

// mapperは使わない。newtaskを使う
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
	dueoption, err := NewDueOption(cmd.DueDate)
	//int->Dueoption。これをhandlerでやってusecaseはdtask.dueoptionにすべきか議論
	if err != nil {
		return Result{}, err
	}
	due, err := dtask.NewDueDateFromOption(now, dueoption)
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

func NewDueOption(t int) (dtask.DueOption, error) {
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

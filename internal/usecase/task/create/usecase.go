package create

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/clock"
	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
)

type Usecase struct {
	tx    usetx.Runner[usetx.TaskEventDeps]
	clock clock.Clocker
}

func New(tx usetx.Runner[usetx.TaskEventDeps], clock clock.Clocker) *Usecase {
	return &Usecase{tx: tx, clock: clock}
}

// mapperは使わない。newtaskを使う
func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	cmd, err := normalize(cmd)
	if err != nil {
		return Result{}, err
	} //usecaseのエラー
	title, err := dtask.NewTaskTitle(cmd.Title)
	if err != nil {
		switch {
		case errors.Is(err, dtask.ErrEmptyTitle):
			return Result{}, ErrEmptyTitle
		case errors.Is(err, dtask.ErrTitleTooLong):
			return Result{}, ErrTitleTooLong
		default:
			return Result{}, err
		}
	}
	desc, err := dtask.NewTaskDescription(cmd.Description)
	if err != nil {
		switch {
		case errors.Is(err, dtask.ErrEmptyDescription):
			return Result{}, ErrEmptyDescription
		case errors.Is(err, dtask.ErrDescriptionTooLong):
			return Result{}, ErrDescriptionTooLong
		default:
			return Result{}, err
		}
	}
	now := u.clock.Now()
	dueoption, err := normalizeDueOption(cmd.DueDate)
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
	userID := user.UserID("tmp") //ここは認証が完成したらuseridをctxから取る
	t := dtask.NewTask(userID, title, desc, due, now)
	//task_id発行はdomainに寄せてるからt.IDで取得する
	//eventType, err := taskevent.ParseEventType("created") parseはinfraで使うようやからここで使うとダサい
	reqID, ok := requestid.FromContext(ctx) //これerrでなくok受け取り
	if !ok || reqID == "" {
		reqID = uuid.NewString()
	}

	//teskevent.RequestID(reqID)これやっていいんや。
	event := taskevent.NewCreatedEvent(
		userID, t.ID(), taskevent.RequestID(reqID), now, taskevent.CreatedPayload{},
	)

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.TaskEventDeps) error {
		if err := deps.TaskRepo().Store(ctx, t); err != nil {
			return err
		}
		if err := deps.TaskEventRepo().Insert(ctx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return Result{}, err
	}

	return Result{ID: t.ID().Value()}, nil
}

func normalizeDueOption(t int) (dtask.DueOption, error) {
	//これ大文字にすると共有になるNormalizeDueOption
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
	} //これはdomainに置くのはよくない。7,14,21,30っていう外部出力に依存してる。

}

package remove

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

func (u *Usecase) Do(ctx context.Context, cmd Command) error {
	//deleteはresult返さずにステータスメッセージだけ
	cmd, err := normalize(cmd)
	if err != nil {
		return err
	}

	id, err := dtask.ParseTaskID(cmd.ID)
	if err != nil {
		return ErrInvalidID
	}

	now := u.clock.Now()
	userID := user.UserID("tmp") //認証完成したらctxから取る

	reqID, ok := requestid.FromContext(ctx)
	if !ok || reqID == "" {
		reqID = uuid.NewString()
	}

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.TaskEventDeps) error {
		t, err := deps.TaskRepo().FindByID(ctx, id)
		if err != nil {
			return err
		}

		dateLeft := int(t.DueDate().Value().Sub(now).Hours() / 24)

		event := taskevent.NewDeletedEvent(
			userID, id, taskevent.RequestID(reqID), now,
			taskevent.DeletedPayload{DateLeft: dateLeft},
		)

		//楽観ロックはrepoに寄せる。
		if err := deps.TaskRepo().Delete(ctx, id, cmd.Version); err != nil {
			return err
		}
		if err := deps.TaskEventRepo().Insert(ctx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		switch {
		case errors.Is(err, dtask.ErrNotFound):
			return ErrNotFound
		case errors.Is(err, dtask.ErrConflict):
			return ErrConflict
		default:
			return err
		}
	}

	return nil
}

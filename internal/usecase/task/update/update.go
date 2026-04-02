package update

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
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("usecase/task/update")

type Usecase struct {
	tx    usetx.Runner[usetx.TaskEventDeps]
	clock clock.Clocker
}

func New(tx usetx.Runner[usetx.TaskEventDeps], clock clock.Clocker) *Usecase {
	return &Usecase{tx: tx, clock: clock}
}
func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	ctx, span := tracer.Start(ctx, "task.update")
	defer span.End()

	cmd, err := normalize(cmd)

	if err != nil {
		return Result{}, err

	}
	id, err := dtask.ParseTaskID(cmd.ID)
	if err != nil {
		return Result{}, ErrInvalidID
	}

	now := u.clock.Now()
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		return Result{}, ErrInvalidID
	}

	reqID, ok := requestid.FromContext(ctx)
	if !ok || reqID == "" {
		reqID = uuid.NewString()
	}

	var fields []taskevent.UpdatedFields
	if cmd.Title != nil {
		fields = append(fields, taskevent.FieldTitle)
	}
	if cmd.Description != nil {
		fields = append(fields, taskevent.FieldDescription)
	}
	if cmd.DueDate != nil {
		fields = append(fields, taskevent.FieldDueDate)
	}

	event := taskevent.NewUpdatedEvent(
		userID, id, taskevent.RequestID(reqID), now,
		taskevent.UpdatedPayload{Fields: fields},
	)

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.TaskEventDeps) error {
		t, err := deps.TaskRepo().FindByID(ctx, id)
		if err != nil {
			return err
		}

		if t.UserID() != userID {
			return dtask.ErrNotFound
		}

		if t.Version() != cmd.Version {

			return dtask.ErrConflict
		}

		if cmd.Title != nil {
			title, err := dtask.NewTaskTitle(*cmd.Title)
			if err != nil {
				return err
			}
			t.ChangeTitle(title, now)
		}
		if cmd.Description != nil {
			desc, err := dtask.NewTaskDescription(*cmd.Description)
			if err != nil {
				return err
			}
			t.ChangeDescription(desc, now)
		}
		if cmd.DueDate != nil {
			opt, err := normalizeDueOption(*cmd.DueDate)
			if err != nil {
				return err
			}
			due, err := dtask.NewDueDateFromOption(now, opt)
			if err != nil {
				return err
			}
			t.ChangeDueDate(due, now)
		}
		if err := deps.TaskRepo().Update(ctx, t); err != nil {
			return err
		}
		if err := deps.TaskEventRepo().Insert(ctx, event); err != nil {
			return err
		}
		return nil
	}); err != nil {
		switch {
		case errors.Is(err, dtask.ErrNotFound):
			return Result{}, ErrNotFound
		case errors.Is(err, dtask.ErrConflict):
			return Result{}, ErrConflict
		case errors.Is(err, dtask.ErrEmptyTitle):
			return Result{}, ErrEmptyTitle
		case errors.Is(err, dtask.ErrTitleTooLong):
			return Result{}, ErrTitleTooLong
		case errors.Is(err, dtask.ErrEmptyDescription):
			return Result{}, ErrEmptyDescription
		case errors.Is(err, dtask.ErrDescriptionTooLong):
			return Result{}, ErrDescriptionTooLong
		default:
			return Result{}, err
		}
	}

	return Result{
		ID: id.Value(),
		//TODO:versionも返すべき。その場合repoに+の責務を寄せてるから更新後selectが必須。
	}, nil
}

func normalizeDueOption(t int) (dtask.DueOption, error) { //Dueはusecaseに寄せる
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
		return 0, ErrInvalidDueOption
	}

}

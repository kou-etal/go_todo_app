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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	//usecaseがこれimportするのは可。そもそもこれがotelの想定パターン。
	//interface作ってDIで受け取るのは過剰。
)

var tracer = otel.Tracer("usecase/task/create")

type Usecase struct {
	tx    usetx.Runner[usetx.TaskEventDeps]
	clock clock.Clocker
}

func New(tx usetx.Runner[usetx.TaskEventDeps], clock clock.Clocker) *Usecase {
	return &Usecase{tx: tx, clock: clock}
}

// mapperは使わない。newtaskを使う
func (u *Usecase) Do(ctx context.Context, cmd Command) (Result, error) {
	ctx, span := tracer.Start(ctx, "task.create")
	defer span.End()

	cmd, err := normalize(cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	} //usecaseのエラー
	title, err := dtask.NewTaskTitle(cmd.Title)
	if err != nil {
		switch {
		case errors.Is(err, dtask.ErrEmptyTitle):
			span.RecordError(ErrEmptyTitle)
			span.SetStatus(codes.Error, ErrEmptyTitle.Error())
			return Result{}, ErrEmptyTitle
		case errors.Is(err, dtask.ErrTitleTooLong):
			span.RecordError(ErrTitleTooLong)
			span.SetStatus(codes.Error, ErrTitleTooLong.Error())
			return Result{}, ErrTitleTooLong
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}
	desc, err := dtask.NewTaskDescription(cmd.Description)
	if err != nil {
		switch {
		case errors.Is(err, dtask.ErrEmptyDescription):
			span.RecordError(ErrEmptyDescription)
			span.SetStatus(codes.Error, ErrEmptyDescription.Error())
			return Result{}, ErrEmptyDescription
		case errors.Is(err, dtask.ErrDescriptionTooLong):
			span.RecordError(ErrDescriptionTooLong)
			span.SetStatus(codes.Error, ErrDescriptionTooLong.Error())
			return Result{}, ErrDescriptionTooLong
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}
	now := u.clock.Now()
	dueoption, err := normalizeDueOption(cmd.DueDate)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}
	due, err := dtask.NewDueDateFromOption(now, dueoption)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		span.RecordError(ErrInvalidUserID)
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		return Result{}, ErrInvalidUserID
	}

	span.SetAttributes(attribute.String("user.id", cmd.UserID))

	t := dtask.NewTask(userID, title, desc, due, now)

	reqID, ok := requestid.FromContext(ctx)
	if !ok || reqID == "" {
		reqID = uuid.NewString()
	}

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
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	return Result{ID: t.ID().Value()}, nil
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
		return 0, ErrInvalidDueOption
	}

}

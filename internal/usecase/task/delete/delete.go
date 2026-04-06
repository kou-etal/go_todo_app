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
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("usecase/task/delete")

type Usecase struct {
	tx    usetx.Runner[usetx.TaskEventDeps]
	clock clock.Clocker
}

func New(tx usetx.Runner[usetx.TaskEventDeps], clock clock.Clocker) *Usecase {
	return &Usecase{tx: tx, clock: clock}
}

func (u *Usecase) Do(ctx context.Context, cmd Command) error {
	ctx, span := tracer.Start(ctx, "task.delete")
	defer span.End()

	cmd, err := normalize(cmd)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	id, err := dtask.ParseTaskID(cmd.ID)
	if err != nil {
		span.RecordError(ErrInvalidID)
		span.SetStatus(codes.Error, ErrInvalidID.Error())
		return ErrInvalidID
	}

	span.SetAttributes(attribute.String("task.id", cmd.ID))

	now := u.clock.Now()
	userID, err := user.ParseUserID(cmd.UserID)
	if err != nil {
		span.RecordError(ErrInvalidID)
		span.SetStatus(codes.Error, ErrInvalidID.Error())
		return ErrInvalidID
	}

	span.SetAttributes(attribute.String("user.id", cmd.UserID))

	reqID, ok := requestid.FromContext(ctx)
	if !ok || reqID == "" {
		reqID = uuid.NewString()
	}

	if err := u.tx.WithinTx(ctx, func(ctx context.Context, deps usetx.TaskEventDeps) error {
		t, err := deps.TaskRepo().FindByID(ctx, id)
		if err != nil {
			return err
		}

		if t.UserID() != userID {
			return dtask.ErrNotFound
		}

		dateLeft := int(t.DueDate().Value().Sub(now).Hours() / 24)

		event := taskevent.NewDeletedEvent(
			userID, id, taskevent.RequestID(reqID), now,
			taskevent.DeletedPayload{DateLeft: dateLeft},
		)

		if err := deps.TaskEventRepo().Insert(ctx, event); err != nil {
			return err
		}
		//楽観ロックはrepoに寄せる。
		if err := deps.TaskRepo().Delete(ctx, id, cmd.Version); err != nil {
			return err
		}
		return nil
	}); err != nil {
		switch {
		case errors.Is(err, dtask.ErrNotFound):
			span.RecordError(ErrNotFound)
			span.SetStatus(codes.Error, ErrNotFound.Error())
			return ErrNotFound
		case errors.Is(err, dtask.ErrConflict):
			span.RecordError(ErrConflict)
			span.SetStatus(codes.Error, ErrConflict.Error())
			return ErrConflict
		default:
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}

	return nil
}

package list

import (
	"context"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

var tracer = otel.Tracer("usecase/task/list")

type Usecase struct {
	repo dtask.TaskRepository
}

func New(repo dtask.TaskRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Do(ctx context.Context, q Query) (Result, error) {
	ctx, span := tracer.Start(ctx, "task.list")
	defer span.End()

	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		span.RecordError(ErrInvalidUserID)
		span.SetStatus(codes.Error, ErrInvalidUserID.Error())
		return Result{}, ErrInvalidUserID
	}

	span.SetAttributes(attribute.String("user.id", q.UserID))

	limit, err := normalizeLimit(q.Limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	sort, err := normalizeSort(q.Sort)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	var cursor *dtask.ListCursor
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err

		}
		cursor = &c
	}

	dq := dtask.ListQuery{
		UserID: userID,
		Limit:  limit,
		Sort:   sort,
		Cursor: cursor,
	}

	tasks, next, err := u.repo.List(ctx, dq)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return Result{}, err
	}

	items := make([]Item, 0, len(tasks))
	for _, t := range tasks {
		items = append(items, mapTaskToItem(t))
	}

	nextCursor := ""

	if next != nil {
		nextCursor, err = encodeCursor(*next)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return Result{}, err
		}
	}

	span.SetAttributes(attribute.Int("task.list.count", len(tasks)))

	return Result{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

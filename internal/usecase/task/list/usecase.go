package list

import (
	"context"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

type Usecase struct {
	repo dtask.TaskRepository
}

func New(repo dtask.TaskRepository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Do(ctx context.Context, q Query) (Result, error) {
	userID, err := user.ParseUserID(q.UserID)
	if err != nil {
		return Result{}, ErrInvalidUserID
	}

	limit, err := normalizeLimit(q.Limit)
	if err != nil {
		return Result{}, err
	}

	sort, err := normalizeSort(q.Sort)
	if err != nil {
		return Result{}, err
	}

	var cursor *dtask.ListCursor
	if q.Cursor != "" {
		c, err := decodeCursor(q.Cursor)
		if err != nil {
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
			return Result{}, err
		}
	}

	return Result{
		Items:      items,
		NextCursor: nextCursor,
	}, nil
}

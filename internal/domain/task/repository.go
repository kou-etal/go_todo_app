package task

import (
	"context"
	"time"

	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

type ListSort string

const (
	SortCreated ListSort = "created"
	SortDueDate ListSort = "dueDate"
)

type ListCursor struct {
	Created   time.Time
	DueIsNull bool
	DueDate   time.Time

	ID TaskID
}

type ListQuery struct {
	UserID user.UserID
	Limit  int
	Sort   ListSort
	Cursor *ListCursor
}

type TaskRepository interface {
	List(ctx context.Context, q ListQuery) ([]*Task, *ListCursor, error)
	Store(ctx context.Context, t *Task) error
	Update(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, id TaskID) (*Task, error)
	Delete(ctx context.Context, id TaskID, userID user.UserID, version uint64) error
}

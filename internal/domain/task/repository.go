package task

import (
	"context"
	"time"
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
	Limit  int
	Sort   ListSort
	Cursor *ListCursor
}

type TaskRepository interface {
	List(ctx context.Context, q ListQuery) ([]*Task, *ListCursor, error)
	Store(ctx context.Context, t *Task) error
	Update(ctx context.Context, t *Task) error
	FindByID(ctx context.Context, id TaskID) (*Task, error)
} //repositoryはinterface定義だけでqueryとかはlist.goに分けるべき

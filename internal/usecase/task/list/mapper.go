package list

import (
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func mapTaskToItem(t *dtask.Task) Item {
	duePtr := toTimePtr(t.DueDate().Value())
	return Item{
		ID:          t.ID().Value(),
		Title:       t.Title().Value(),
		Description: t.Description().Value(),
		Status:      t.Status().Value(),
		DueDate:     duePtr,
	}
}

func toTimePtr(v time.Time) *time.Time {
	x := v
	return &x
}

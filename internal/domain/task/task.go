package task

import (
	"time"

	"github.com/kou-etal/go_todo_app/internal/clock"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

type Task struct {
	id          TaskID
	userID      user.UserID
	title       TaskTitle
	description TaskDescription
	status      TaskStatus
	dueDate     DueDate
	createdAt   time.Time
	updatedAt   time.Time
	version     uint64
}

type Tasks []*Task

func (t *Task) ID() TaskID                   { return t.id }
func (t *Task) UserID() user.UserID          { return t.userID }
func (t *Task) Title() TaskTitle             { return t.title }
func (t *Task) Description() TaskDescription { return t.description }
func (t *Task) Status() TaskStatus           { return t.status }
func (t *Task) DueDate() DueDate             { return t.dueDate }
func (t *Task) CreatedAt() time.Time         { return t.createdAt }
func (t *Task) UpdatedAt() time.Time         { return t.updatedAt }
func (t *Task) Version() uint64              { return t.version }

func (t *Task) ChangeTitle(newTitle TaskTitle, now time.Time) {
	t.title = newTitle
	t.updateTime(now)
}
func (t *Task) ChangeDescription(newDesc TaskDescription, now time.Time) {
	t.description = newDesc
	t.updateTime(now)
}
func (t *Task) ChangeDueDate(newDue DueDate, now time.Time) {
	t.dueDate = newDue
	t.updateTime(now)
}

func (t *Task) ChangeStatus(next TaskStatus, now time.Time) error {
	if t.status == StatusDone {
		return ErrStatusChangeDone
	}

	t.status = next
	t.updateTime(now)
	return nil
}

func (t *Task) Reschedule(newDue DueDate, now time.Time) {

	t.dueDate = newDue
	t.updateTime(now)
}

func (t *Task) MarkDone(now time.Time) error {
	return t.ChangeStatus(StatusDone, now)
}

func (t *Task) updateTime(now time.Time) {
	n := clock.NormalizeTime(now)
	t.updatedAt = n
}

package task

import (
	"errors"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		panic(err)
	}
	return t
}

func reconstructTestTask(status TaskStatus) *Task {
	title, _ := NewTaskTitle("test title")
	desc, _ := NewTaskDescription("test description")
	due, _ := NewDueDateFromTime(time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC))
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return ReconstructTask(
		NewTaskID(), user.UserID("u-1"),
		title, desc, status, due,
		created, created, 1,
	)
}

func TestChangeStatus_fromTodo(t *testing.T) {
	t.Parallel()

	task := reconstructTestTask(StatusTodo)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	if err := task.ChangeStatus(StatusDoing, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status() != StatusDoing {
		t.Fatalf("status = %v, want doing", task.Status())
	}
	if task.UpdatedAt().Before(task.CreatedAt()) {
		t.Fatal("updatedAt should be updated")
	}
}

func TestChangeStatus_fromDone(t *testing.T) {
	t.Parallel()

	task := reconstructTestTask(StatusDone)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	err := task.ChangeStatus(StatusTodo, now)
	if !errors.Is(err, ErrStatusChangeDone) {
		t.Fatalf("err = %v, want ErrStatusChangeDone", err)
	}
}

func TestMarkDone(t *testing.T) {
	t.Parallel()

	task := reconstructTestTask(StatusTodo)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	if err := task.MarkDone(now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Status() != StatusDone {
		t.Fatalf("status = %v, want done", task.Status())
	}
}

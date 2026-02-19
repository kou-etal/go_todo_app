package task

import (
	"errors"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

func TestNewTask(t *testing.T) {
	t.Parallel()

	title, _ := NewTaskTitle("my task")
	desc, _ := NewTaskDescription("my desc")
	now := time.Date(2026, 2, 1, 10, 30, 0, 0, time.UTC)
	due, _ := NewDueDateFromTime(now.AddDate(0, 0, 7))

	task := NewTask(user.UserID("u-1"), title, desc, due, now)

	if task.ID().Value() == "" {
		t.Fatal("ID should not be empty")
	}
	if task.Status() != StatusTodo {
		t.Fatalf("status = %v, want todo", task.Status())
	}
	if task.Version() != 1 {
		t.Fatalf("version = %d, want 1", task.Version())
	}
	if !task.CreatedAt().Equal(task.UpdatedAt()) {
		t.Fatal("createdAt should equal updatedAt")
	}
}

func TestNewDueDateFromOption(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		opt     DueOption
		wantErr error
	}{
		{"7_days", Due7Days, nil},
		{"14_days", Due14Days, nil},
		{"21_days", Due21Days, nil},
		{"30_days", Due30Days, nil},
		{"zero", DueOption(0), ErrInvalidDueOption},
		{"negative", DueOption(-1), ErrInvalidDueOption},
		{"over_30", DueOption(31), ErrInvalidDueOption},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewDueDateFromOption(now, tt.opt)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			want := now.AddDate(0, 0, int(tt.opt))
			if !got.Value().Equal(want) {
				t.Fatalf("due = %v, want %v", got.Value(), want)
			}
		})
	}
}

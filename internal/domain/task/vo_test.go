package task

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNewTaskTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid", "buy groceries", nil},
		{"empty", "", ErrEmptyTitle},
		{"whitespace_only", "   ", ErrEmptyTitle},
		{"exactly_20_runes", strings.Repeat("a", 20), nil},
		{"21_runes_too_long", strings.Repeat("a", 21), ErrTitleTooLong},
		{"multibyte_20_runes_ok", strings.Repeat("あ", 20), nil},
		{"multibyte_21_runes_ng", strings.Repeat("あ", 21), ErrTitleTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewTaskTitle(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value() != tt.input {
				t.Fatalf("value = %q, want %q", got.Value(), tt.input)
			}
		})
	}
}

func TestNewTaskDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid", "some description", nil},
		{"empty", "", ErrEmptyDescription},
		{"whitespace_only", "   ", ErrEmptyDescription},
		{"exactly_1000_runes", strings.Repeat("a", 1000), nil},
		{"1001_runes_too_long", strings.Repeat("a", 1001), ErrDescriptionTooLong},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewTaskDescription(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value() != tt.input {
				t.Fatalf("value = %q, want %q", got.Value(), tt.input)
			}
		})
	}
}

func TestParseTaskStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    TaskStatus
		wantErr error
	}{
		{"todo", "todo", StatusTodo, nil},
		{"doing", "doing", StatusDoing, nil},
		{"done", "done", StatusDone, nil},
		{"invalid", "archived", "", ErrInvalidStatus},
		{"empty", "", "", ErrInvalidStatus},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTaskStatus(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("status = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseTaskID(t *testing.T) {
	t.Parallel()

	validUUID := uuid.NewString()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"valid_uuid", validUUID, nil},
		{"empty", "", ErrInvalidID},
		{"invalid_format", "not-a-uuid", ErrInvalidID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseTaskID(tt.input)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Value() != tt.input {
				t.Fatalf("value = %q, want %q", got.Value(), tt.input)
			}
		})
	}
}

func TestNewDueDateFromTime(t *testing.T) {
	t.Parallel()

	input := mustParseTime("2026-02-15T10:30:45.123456789Z")
	got, err := NewDueDateFromTime(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v := got.Value()
	if !v.Equal(mustParseTime("2026-02-15T10:30:45Z")) {
		t.Fatalf("value = %v, want truncated to second", v)
	}
	if v.Location().String() != "UTC" {
		t.Fatalf("location = %v, want UTC", v.Location())
	}
}

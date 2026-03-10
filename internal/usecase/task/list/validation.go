package list

import (
	"errors"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

var (
	ErrInvalidUserID = errors.New("invalid user id")
	ErrInvalidLimit  = errors.New("invalid limit")
	ErrInvalidSort   = errors.New("invalid sort")
	ErrInvalidCursor = errors.New("invalid cursor")
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

func normalizeLimit(v int) (int, error) {
	if v == 0 {
		return defaultLimit, nil
	}

	if v < 1 || v > maxLimit {
		return 0, ErrInvalidLimit
	}
	return v, nil
}

func normalizeSort(v string) (dtask.ListSort, error) {
	switch v {
	case "", "created":
		return dtask.SortCreated, nil
	case "dueDate":
		return dtask.SortDueDate, nil
	default:
		return "", ErrInvalidSort
	}
}

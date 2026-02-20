package task

import (
	"strings"
	"unicode/utf8"
)

type TaskTitle struct {
	value string
}

func NewTaskTitle(v string) (TaskTitle, error) {
	if strings.TrimSpace(v) == "" {
		return TaskTitle{}, ErrEmptyTitle
	}
	if utf8.RuneCountInString(v) > 20 {
		return TaskTitle{}, ErrTitleTooLong
	}

	return TaskTitle{value: v}, nil
}

func (t TaskTitle) Value() string {
	return t.value
}

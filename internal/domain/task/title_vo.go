package task

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type TaskTitle struct {
	value string
}

func NewTaskTitle(v string) (TaskTitle, error) {
	if strings.TrimSpace(v) == "" {
		return TaskTitle{}, errors.New("tmp")
	}
	if utf8.RuneCountInString(v) > 20 {
		return TaskTitle{}, errors.New("tmp")
	}

	return TaskTitle{value: v}, nil
}

func (t TaskTitle) Value() string {
	return t.value
}

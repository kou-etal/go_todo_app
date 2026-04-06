package task

import (
	"strings"
	"unicode/utf8"
)

type TaskDescription struct {
	value string
}

func NewTaskDescription(v string) (TaskDescription, error) {
	if strings.TrimSpace(v) == "" {
		return TaskDescription{}, ErrEmptyDescription
	}
	if utf8.RuneCountInString(v) > 1000 {
		return TaskDescription{}, ErrDescriptionTooLong
	}

	return TaskDescription{value: v}, nil
}

func (t TaskDescription) Value() string {
	return t.value
}

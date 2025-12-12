package task

import (
	"errors"
	"strings"
	"unicode/utf8"
)

type TaskDescription struct {
	value string
}

func NewTaskDescription(v string) (TaskDescription, error) {
	//TODO:"  abc  "をどうするか  vv:=strings.TrimSpace(v) utf8.RuneCountInString(vv)
	if strings.TrimSpace(v) == "" {
		return TaskDescription{}, errors.New("tmp")
	}
	if utf8.RuneCountInString(v) > 1000 {
		return TaskDescription{}, errors.New("tmp")
	}

	return TaskDescription{value: v}, nil
}

func (t TaskDescription) Value() string {
	return t.value
}

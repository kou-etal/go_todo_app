package task

import (
	"errors"

	"github.com/google/uuid"
)

type TaskID string

// ULIDも存在、より便利
func NewTaskID() TaskID {
	return TaskID(uuid.New().String())
}

func ParseTaskID(v string) (TaskID, error) {
	_, err := uuid.Parse(v)
	if err != nil {
		return "", errors.New("tmp")
	}
	return TaskID(v), nil
}
func (id TaskID) Value() string {
	return string(id)
}

package task

import "errors"

type TaskStatus string

const (
	StatusTodo  TaskStatus = "todo"
	StatusDoing TaskStatus = "doing"
	StatusDone  TaskStatus = "done"
)

func ParseTaskStatus(v string) (TaskStatus, error) {
	switch v {
	case string(StatusTodo):
		return StatusTodo, nil
	case string(StatusDoing):
		return StatusDoing, nil
	case string(StatusDone):
		return StatusDone, nil
	default:
		return "", errors.New("tmp")
	}
}

func (t TaskStatus) Value() string {
	return string(t)
} //これは抽象化、string(t)は表現

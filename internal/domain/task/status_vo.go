package task

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
		return "", ErrInvalidStatus
	}
}

func (t TaskStatus) Value() string {
	return string(t)
}

package task

import "errors"

type TaskStatus string

//statusはdomainのルール。ゆえに宣言がdomain層で可能。簡単に書ける。
//もしstatusがUI都合ならばここで宣言不可、ゆえに境界をpresentation/usecaseに置かなければならない(UI都合をここで受け取れない)。だるい。
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

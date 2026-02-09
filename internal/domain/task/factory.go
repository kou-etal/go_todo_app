package task

import (
	"time"
)

func NewTask(
	title TaskTitle,
	description TaskDescription,
	dueDate DueDate,
	now time.Time,
) *Task { //Entityはコピーされない。ポインタで返す
	n := normalizeTime(now)

	return &Task{
		id:          NewTaskID(), //IDが欲しいのはdomain層の都合
		title:       title,
		description: description,
		status:      StatusTodo,
		dueDate:     dueDate,
		createdAt:   n,
		updatedAt:   n,
		version:     1,
	}
}

// これは復元用。repoで使う
func ReconstructTask(
	id TaskID,
	title TaskTitle,
	description TaskDescription,
	status TaskStatus,
	dueDate DueDate,
	createdAt time.Time,
	updatedAt time.Time,
	version uint64,
) *Task {
	return &Task{
		id:          id,
		title:       title,
		description: description,
		status:      status,
		dueDate:     dueDate,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		version:     version,
	}
}

const maxDueDays = 30 //気遣い
// これは相対、newのロジック->factoryに置く
// ただの相対の状態遷移ロジックはentityに置く
// 相対ではないnewのロジック
func NewDueDateFromOption(now time.Time, opt DueOption) (DueDate, error) {
	now = normalizeTime(now) //factory側で秒

	if opt <= 0 {
		return DueDate{}, ErrInvalidDueOption
	}
	//TODO:キャストするの良くないらしい
	if int(opt) > maxDueDays {
		return DueDate{}, ErrInvalidDueOption
	}
	target := now.AddDate(0, 0, int(opt))

	if target.Before(now) {
		return DueDate{}, ErrInvalidDueOption
	} //保険
	//TODO:保険でも読む側に負担与えるならば良くない

	return NewDueDateFromTime(target)
}

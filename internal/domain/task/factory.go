package task

import (
	"errors"
	"time"
)

func NewTask(
	title TaskTitle,
	description TaskDescription,
	dueDate DueDate,
	now time.Time,
) *Task {
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

const maxDueDays = 365 //気遣い

func NewDueDateFromOption(now time.Time, opt DueOption) (DueDate, error) {
	now = normalizeTime(now) //factory側で秒

	if opt <= 0 {
		return DueDate{}, errors.New("tmp")
	}
	//TODO:キャストするの良くないらしい
	if int(opt) > maxDueDays {
		return DueDate{}, errors.New("tmp")
	}

	target := now.AddDate(0, 0, int(opt))

	if target.Before(now) {
		return DueDate{}, errors.New("tmp")
	} //保険
	//TODO:保険でも読む側に負担与えるならば良くない

	return NewDueDateFromTime(target)
}

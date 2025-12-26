package task

import "time"

type DueDate struct {
	value time.Time
}

func NewDueDateFromTime(t time.Time) (DueDate, error) {
	t = t.UTC().Truncate(time.Second)
	return DueDate{value: t}, nil
}
func (d DueDate) Value() time.Time {
	return d.value
}

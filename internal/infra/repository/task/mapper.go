package taskrepo

import (
	"fmt"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
)

func RecordToEntity(r *TaskRecord) (*dtask.Task, error) {

	id := dtask.TaskID(r.ID)

	title, err := dtask.NewTaskTitle(r.Title)
	if err != nil {
		return nil, fmt.Errorf("invalid task record id=%s field=title: %w", r.ID, err)
	} //これはDBがバグってることによるエラーゆえにdomainのエラーは使わずにwrapして返す

	description, err := dtask.NewTaskDescription(r.Description)
	if err != nil {
		return nil, fmt.Errorf("invalid task record id=%s field=description: %w", r.ID, err)
	}

	status, err := dtask.ParseTaskStatus(r.Status)
	if err != nil {
		return nil, fmt.Errorf("invalid task record id=%s field=status value=%q: %w", r.ID, r.Status, err)
	}

	due, err := dtask.NewDueDateFromTime(r.DueDate)
	if err != nil {
		return nil, fmt.Errorf("invalid task record id=%s field=dueDate: %w", r.ID, err)
	}

	return dtask.ReconstructTask(
		id,
		title,
		description,
		status,
		due,
		r.Created,
		r.Updated,
		r.Version,
	), nil
}

func EntityToRecord(t *dtask.Task) *TaskRecord {

	return &TaskRecord{
		ID:          string(t.ID()),
		Title:       string(t.Title().Value()),
		Description: string(t.Description().Value()),
		Status:      string(t.Status()),
		DueDate:     time.Time(t.DueDate().Value()),
		Created:     t.CreatedAt(),
		Updated:     t.UpdatedAt(),
		Version:     t.Version(),
	}
}

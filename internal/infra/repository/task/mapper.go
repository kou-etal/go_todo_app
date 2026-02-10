package taskrepo

import (
	"fmt"
	"time"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
)

func RecordToEntity(r *TaskRecord) (*dtask.Task, error) {

	id := dtask.TaskID(r.ID) //TODO:ちゃんとparseidで検証しよう
	userID := duser.UserID(r.UserID)

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
	/*ReconstructTaskを作らなければtask := &Task{
	      id: id,
	      title: title,
	      ...
	  	となりrepoがdomainに関与しすぎる
	*/

	return dtask.ReconstructTask(
		id,
		userID,
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
		ID:          t.ID().Value(), //これdomainでstringにキャストして返してるのにもっかいstringは冗長。
		UserID:      t.UserID().Value(),
		Title:       t.Title().Value(),
		Description: t.Description().Value(),
		Status:      t.Status().Value(),
		DueDate:     time.Time(t.DueDate().Value()),
		Created:     t.CreatedAt(),
		Updated:     t.UpdatedAt(),
		Version:     t.Version(),
	}
}

//外から中(record->entity)は疑う。中から外(entity->record)は信用する

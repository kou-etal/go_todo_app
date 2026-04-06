package taskeventrepo

import (
	"encoding/json"
	"fmt"

	taskevent "github.com/kou-etal/go_todo_app/internal/domain/event"
)

func toRecord(e *taskevent.TaskEvent) (TaskEventRecord, error) {

	p, err := json.Marshal(e.Payload())
	if err != nil {
		return TaskEventRecord{}, fmt.Errorf("marshal payload: %w", err)
	}
	return TaskEventRecord{
		ID:            e.ID().Value(),
		UserID:        e.UserID().Value(),
		TaskID:        e.TaskID().Value(),
		RequestID:     e.RequestID().Value(),
		EventType:     e.EventType().Value(),
		OccurredAt:    e.OccurredAt(),
		EmittedAt:     nil,
		SchemaVersion: 1,
		Payload:       json.RawMessage(p),
	}, nil
}

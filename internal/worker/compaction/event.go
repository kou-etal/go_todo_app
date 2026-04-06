package compaction

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID            string          `json:"id"`
	UserID        string          `json:"user_id"`
	TaskID        string          `json:"task_id"`
	RequestID     string          `json:"request_id"`
	EventType     string          `json:"event_type"`
	OccurredAt    time.Time       `json:"occurred_at"`
	SchemaVersion uint32          `json:"schema_version"`
	Payload       json.RawMessage `json:"payload"`
}

func (e *Event) OccurredDay() string {
	return e.OccurredAt.UTC().Format("2006-01-02")

}

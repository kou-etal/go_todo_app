package taskeventrepo

import (
	"encoding/json"
	"time"
)

type TaskEventRecord struct {
	ID            string          `db:"id"`
	UserID        string          `db:"user_id"`
	TaskID        string          `db:"task_id"`
	RequestID     string          `db:"request_id"`
	EventType     string          `db:"event_type"`
	OccurredAt    time.Time       `db:"occurred_at"`
	EmittedAt     *time.Time      `db:"emitted_at"`
	SchemaVersion uint32          `db:"schema_version"`
	Payload       json.RawMessage `db:"payload"`
	NextAttemptAt time.Time       `db:"next_attempt_at"`
	AttemptCount  uint32          `db:"attempt_count"`
	LeaseOwner    *string         `db:"lease_owner"`
	LeaseUntil    *time.Time      `db:"lease_until"`
	ClaimedAt     *time.Time      `db:"claimed_at"`
}

type TaskEventDLQRecord struct {
	ID            string          `db:"id"`
	UserID        string          `db:"user_id"`
	TaskID        string          `db:"task_id"`
	RequestID     string          `db:"request_id"`
	EventType     string          `db:"event_type"`
	OccurredAt    time.Time       `db:"occurred_at"`
	SchemaVersion uint32          `db:"schema_version"`
	Payload       json.RawMessage `db:"payload"`
	AttemptCount  uint32          `db:"attempt_count"`
	LastError     *string         `db:"last_error"`
	DeadAt        time.Time       `db:"dead_at"`
}

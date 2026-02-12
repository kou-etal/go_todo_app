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
	EmittedAt     *time.Time      `db:"emitted_at"`     //nil使う
	SchemaVersion uint32          `db:"schema_version"` //floatにしたくなるけど誤差で終わる。uint64もいらん
	Payload       json.RawMessage `db:"payload"`        //ここはGo側でjsonに変換。anyで保存できない。
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
	SchemaVersion uint32          `db:"schema_version"` //これは必須 version1はこういう操作。
	Payload       json.RawMessage `db:"payload"`        //DB側でJSON
	AttemptCount  uint32          `db:"attempt_count"`
	LastError     *string         `db:"last_error"` //エラー出ずに詰む場合もある。それをundefinedとか入れたら良くない。null許容。
	DeadAt        time.Time       `db:"dead_at"`
}

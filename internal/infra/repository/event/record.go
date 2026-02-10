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
}

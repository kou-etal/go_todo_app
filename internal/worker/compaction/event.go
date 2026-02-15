package compaction

import (
	"encoding/json"
	"time"
)

// 責務が違うからoutboxのeventと分ける。
// Event は jsonl から読み込んだイベント。reader / deduper / writer 間で共有する。
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

// OccurredDay は occurred_at の日付部分を返す（パーティションキー用）。
func (e *Event) OccurredDay() string {
	return e.OccurredAt.UTC().Format("2006-01-02")
	//Goの記法。YYYY-MM-DDプレースホルダーではなく("2006-01-02")で記述。
}

package compaction

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWriteParquet_producesNonEmptyOutput(t *testing.T) {
	t.Parallel()

	events := []Event{
		{
			ID: "e1", UserID: "u1", TaskID: "t1", RequestID: "r1",
			EventType: "created",
			OccurredAt: time.Date(2026, 2, 11, 14, 0, 0, 0, time.UTC),
			SchemaVersion: 1,
			Payload: json.RawMessage(`{}`),
		},
		{
			ID: "e2", UserID: "u2", TaskID: "t2", RequestID: "r2",
			EventType: "updated",
			OccurredAt: time.Date(2026, 2, 11, 15, 0, 0, 0, time.UTC),
			SchemaVersion: 1,
			Payload: json.RawMessage(`{"fields":["title"]}`),
		},
	}

	data, err := writeParquet(events)
	if err != nil {
		t.Fatalf("writeParquet() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("writeParquet() produced empty output")
	}
	// Parquet magic bytes: "PAR1"
	if string(data[:4]) != "PAR1" {
		t.Fatalf("output does not start with PAR1 magic: %x", data[:4])
	}
}

func TestWriteParquet_emptyEvents(t *testing.T) {
	t.Parallel()

	data, err := writeParquet(nil)
	if err != nil {
		t.Fatalf("writeParquet(nil) error: %v", err)
	}
	// 空でも valid な Parquet ファイルが生成される
	if len(data) == 0 {
		t.Fatal("writeParquet(nil) produced empty output")
	}
}

func TestCompactedKey_format(t *testing.T) {
	t.Parallel()

	key := compactedKey("compacted/task-events", "2026-02-11", "2026-02-12")
	want := "compacted/task-events/day=2026-02-11/from-claims-2026-02-12.parquet"
	if key != want {
		t.Fatalf("compactedKey = %s, want %s", key, want)
	}
}

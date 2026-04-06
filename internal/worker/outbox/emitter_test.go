package outbox

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestBuildJSONLines_outputsOneLinePerRecord(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	records := []ClaimedEvent{
		{
			ID: "e1", UserID: "u1", TaskID: "t1", RequestID: "r1",
			EventType: "created", OccurredAt: now, SchemaVersion: 1,
			Payload: json.RawMessage(`{}`),
		},
		{
			ID: "e2", UserID: "u2", TaskID: "t2", RequestID: "r2",
			EventType: "updated", OccurredAt: now, SchemaVersion: 1,
			Payload: json.RawMessage(`{"fields":["title"]}`),
		},
	}

	data, err := buildJSONLines(records)
	if err != nil {
		t.Fatalf("buildJSONLines() error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("lines count = %d, want 2", len(lines))
	}

	// 各行が valid JSON か
	for i, line := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("line %d is not valid JSON: %v", i, err)
		}
	}

	// 1行目に e1 の id が含まれる
	if !strings.Contains(lines[0], `"e1"`) {
		t.Fatalf("line 0 missing id e1: %s", lines[0])
	}
}

func TestBatchID_deterministic(t *testing.T) {
	t.Parallel()

	ids := []string{"id-3", "id-1", "id-2"}
	claimedAt := time.Date(2026, 2, 1, 14, 30, 0, 0, time.UTC)

	b1 := batchID(ids, claimedAt, 1)
	b2 := batchID(ids, claimedAt, 1)

	if b1 != b2 {
		t.Fatalf("batchID not deterministic: %s != %s", b1, b2)
	}
	if len(b1) != 16 {
		t.Fatalf("batchID length = %d, want 16", len(b1))
	}
}

func TestBatchID_orderIndependent(t *testing.T) {
	t.Parallel()

	claimedAt := time.Date(2026, 2, 1, 14, 0, 0, 0, time.UTC)

	b1 := batchID([]string{"a", "b", "c"}, claimedAt, 1)
	b2 := batchID([]string{"c", "a", "b"}, claimedAt, 1)

	if b1 != b2 {
		t.Fatalf("batchID should be order-independent: %s != %s", b1, b2)
	}
}

func TestBatchID_differentInputsDifferentOutput(t *testing.T) {
	t.Parallel()

	claimedAt := time.Date(2026, 2, 1, 14, 0, 0, 0, time.UTC)

	b1 := batchID([]string{"a"}, claimedAt, 1)
	b2 := batchID([]string{"b"}, claimedAt, 1)

	if b1 == b2 {
		t.Fatalf("different inputs should produce different batchID")
	}
}

func TestS3DataKey_hivestyle(t *testing.T) {
	t.Parallel()

	claimedAt := time.Date(2026, 2, 11, 14, 0, 0, 0, time.UTC)
	key := s3DataKey("task-events", claimedAt, "abc123")

	want := "raw/task-events/year=2026/month=02/day=11/hour=14/abc123.jsonl"
	if key != want {
		t.Fatalf("s3DataKey = %s, want %s", key, want)
	}
}

func TestS3ManifestKey_hivestyle(t *testing.T) {
	t.Parallel()

	claimedAt := time.Date(2026, 2, 11, 14, 0, 0, 0, time.UTC)
	key := s3ManifestKey("task-events", claimedAt, "abc123")

	want := "raw/task-events/year=2026/month=02/day=11/hour=14/abc123.manifest.json"
	if key != want {
		t.Fatalf("s3ManifestKey = %s, want %s", key, want)
	}
}

func TestTrimByByteLimit_cutsAtLimit(t *testing.T) {
	t.Parallel()

	records := []ClaimedEvent{
		{ID: "1", Payload: json.RawMessage(strings.Repeat("x", 100))},
		{ID: "2", Payload: json.RawMessage(strings.Repeat("x", 100))},
		{ID: "3", Payload: json.RawMessage(strings.Repeat("x", 100))},
	}

	// 上限 150 → 1件目(100) OK, 2件目(200) 超過 → 1件だけ返る
	result := trimByByteLimit(records, 150)
	if len(result) != 1 {
		t.Fatalf("trimByByteLimit len = %d, want 1", len(result))
	}
}

func TestTrimByByteLimit_allFit(t *testing.T) {
	t.Parallel()

	records := []ClaimedEvent{
		{ID: "1", Payload: json.RawMessage(`{}`)},
		{ID: "2", Payload: json.RawMessage(`{}`)},
	}

	result := trimByByteLimit(records, 1000)
	if len(result) != 2 {
		t.Fatalf("trimByByteLimit len = %d, want 2", len(result))
	}
}

func TestCollectIDs(t *testing.T) {
	t.Parallel()

	records := []ClaimedEvent{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}

	ids := collectIDs(records)
	if len(ids) != 3 {
		t.Fatalf("collectIDs len = %d, want 3", len(ids))
	}
	if ids[0] != "a" || ids[1] != "b" || ids[2] != "c" {
		t.Fatalf("collectIDs = %v, want [a b c]", ids)
	}
}

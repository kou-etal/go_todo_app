package compaction

import (
	"encoding/json"
	"testing"
	"time"
)

func mustEvent(id string, day string) Event {
	t, _ := time.Parse("2006-01-02", day)
	return Event{
		ID:            id,
		UserID:        "u1",
		TaskID:        "t1",
		RequestID:     "r1",
		EventType:     "created",
		OccurredAt:    t,
		SchemaVersion: 1,
		Payload:       json.RawMessage(`{}`),
	}
}

func TestDedupe_removesDuplicates(t *testing.T) {
	t.Parallel()

	events := []Event{
		mustEvent("e1", "2026-02-11"),
		mustEvent("e2", "2026-02-11"),
		mustEvent("e1", "2026-02-11"), // 重複
		mustEvent("e3", "2026-02-11"),
		mustEvent("e2", "2026-02-11"), // 重複
	}

	deduped, removed := dedupe(events)
	if len(deduped) != 3 {
		t.Fatalf("deduped len = %d, want 3", len(deduped))
	}
	if removed != 2 {
		t.Fatalf("removed = %d, want 2", removed)
	}
	// 順序は最初に出現した順
	if deduped[0].ID != "e1" || deduped[1].ID != "e2" || deduped[2].ID != "e3" {
		t.Fatalf("deduped order = %v, want [e1 e2 e3]", []string{deduped[0].ID, deduped[1].ID, deduped[2].ID})
	}
}

func TestDedupe_noDuplicates(t *testing.T) {
	t.Parallel()

	events := []Event{
		mustEvent("e1", "2026-02-11"),
		mustEvent("e2", "2026-02-11"),
	}

	deduped, removed := dedupe(events)
	if len(deduped) != 2 {
		t.Fatalf("deduped len = %d, want 2", len(deduped))
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0", removed)
	}
}

func TestDedupe_empty(t *testing.T) {
	t.Parallel()

	deduped, removed := dedupe(nil)
	if len(deduped) != 0 {
		t.Fatalf("deduped len = %d, want 0", len(deduped))
	}
	if removed != 0 {
		t.Fatalf("removed = %d, want 0", removed)
	}
}

func TestGroupByDay_groupsCorrectly(t *testing.T) {
	t.Parallel()

	events := []Event{
		mustEvent("e1", "2026-02-11"),
		mustEvent("e2", "2026-02-12"),
		mustEvent("e3", "2026-02-11"),
		mustEvent("e4", "2026-02-12"),
		mustEvent("e5", "2026-02-13"),
	}

	groups := groupByDay(events)
	if len(groups) != 3 {
		t.Fatalf("groups count = %d, want 3", len(groups))
	}
	if len(groups["2026-02-11"]) != 2 {
		t.Fatalf("day 2026-02-11 count = %d, want 2", len(groups["2026-02-11"]))
	}
	if len(groups["2026-02-12"]) != 2 {
		t.Fatalf("day 2026-02-12 count = %d, want 2", len(groups["2026-02-12"]))
	}
	if len(groups["2026-02-13"]) != 1 {
		t.Fatalf("day 2026-02-13 count = %d, want 1", len(groups["2026-02-13"]))
	}
}

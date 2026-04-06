package taskeventrepo

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMoveToDLQ_emptyIDs_noop(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)

	err := repo.MoveToDLQ(context.Background(), nil, "some error", time.Now())
	if err != nil {
		t.Fatalf("MoveToDLQ() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 0 {
		t.Fatalf("ExecContext should not be called for empty ids")
	}
}

func TestMoveToDLQ_executesInsertThenDelete(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 1}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	err := repo.MoveToDLQ(context.Background(), []string{"id-1"}, "upload failed", now)
	if err != nil {
		t.Fatalf("MoveToDLQ() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 2 {
		t.Fatalf("ExecContext call count = %d, want 2 (insert + delete)", len(stub.execCalls))
	}

	insertSQL := stub.execCalls[0].sql
	if !strings.Contains(insertSQL, "task_events_dlq") {
		t.Fatalf("first exec should be INSERT into dlq:\n%s", insertSQL)
	}
	if !strings.Contains(insertSQL, "INSERT") {
		t.Fatalf("first exec missing INSERT:\n%s", insertSQL)
	}

	deleteSQL := stub.execCalls[1].sql
	if !strings.Contains(deleteSQL, "DELETE FROM task_events") {
		t.Fatalf("second exec should be DELETE from task_events:\n%s", deleteSQL)
	}
}

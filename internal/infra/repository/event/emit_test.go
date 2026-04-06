package taskeventrepo

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMarkEmitted_emptyIDs_noop(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)

	err := repo.MarkEmitted(context.Background(), nil, "worker-1", time.Now())
	if err != nil {
		t.Fatalf("MarkEmitted() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 0 {
		t.Fatalf("ExecContext should not be called for empty ids")
	}
}

func TestMarkEmitted_buildsCorrectSQL(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 2}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	err := repo.MarkEmitted(context.Background(), []string{"id-1", "id-2"}, "worker-1", now)
	if err != nil {
		t.Fatalf("MarkEmitted() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 1 {
		t.Fatalf("ExecContext call count = %d, want 1", len(stub.execCalls))
	}
	sql := stub.execCalls[0].sql
	if !strings.Contains(sql, "emitted_at") {
		t.Fatalf("sql missing emitted_at:\n%s", sql)
	}
	if !strings.Contains(sql, "lease_owner = NULL") {
		t.Fatalf("sql missing lease_owner clear:\n%s", sql)
	}
}

func TestMarkRetry_emptyIDs_noop(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)

	err := repo.MarkRetry(context.Background(), nil, "worker-1", time.Now())
	if err != nil {
		t.Fatalf("MarkRetry() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 0 {
		t.Fatalf("ExecContext should not be called for empty ids")
	}
}

func TestMarkRetry_buildsCorrectSQL(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 1}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	nextAttempt := now.Add(1 * time.Minute)

	err := repo.MarkRetry(context.Background(), []string{"id-1"}, "worker-1", nextAttempt)
	if err != nil {
		t.Fatalf("MarkRetry() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 1 {
		t.Fatalf("ExecContext call count = %d, want 1", len(stub.execCalls))
	}
	sql := stub.execCalls[0].sql
	if !strings.Contains(sql, "attempt_count = attempt_count + 1") {
		t.Fatalf("sql missing attempt_count increment:\n%s", sql)
	}
	if !strings.Contains(sql, "next_attempt_at") {
		t.Fatalf("sql missing next_attempt_at:\n%s", sql)
	}
	if !strings.Contains(sql, "lease_owner = NULL") {
		t.Fatalf("sql missing lease clear:\n%s", sql)
	}
}

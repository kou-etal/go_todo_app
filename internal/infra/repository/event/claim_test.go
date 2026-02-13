package taskeventrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

// --- stub ---

type stubResult struct {
	affected int64
}

func (s stubResult) LastInsertId() (int64, error) { return 0, nil }
func (s stubResult) RowsAffected() (int64, error) { return s.affected, nil }

type execCall struct {
	sql  string
	args []any
}

type stubQE struct {
	// SelectContext用
	selectRecords []TaskEventRecord
	selectErr     error
	gotSelectSQL  string
	gotSelectArgs []any

	// ExecContext用
	execCalls []execCall
	execErr   error
	execAffected int64
}

func (s *stubQE) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	s.gotSelectSQL = query
	s.gotSelectArgs = args
	if s.selectErr != nil {
		return s.selectErr
	}
	p, ok := dest.(*[]TaskEventRecord)
	if !ok {
		return errors.New("dest is not *[]TaskEventRecord")
	}
	*p = append((*p)[:0], s.selectRecords...)
	return nil
}

func (s *stubQE) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	s.execCalls = append(s.execCalls, execCall{sql: query, args: args})
	if s.execErr != nil {
		return nil, s.execErr
	}
	return stubResult{affected: s.execAffected}, nil
}

func (s *stubQE) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return nil, errors.New("unexpected NamedExecContext call")
}
func (s *stubQE) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return errors.New("unexpected GetContext call")
}
func (s *stubQE) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	return nil, errors.New("unexpected QueryxContext call")
}
func (s *stubQE) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	panic("unexpected QueryRowxContext call")
}
func (s *stubQE) PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error) {
	return nil, errors.New("unexpected PreparexContext call")
}
func (s *stubQE) Rebind(query string) string {
	return query
}

// --- Claim tests ---

func TestClaim_buildsCorrectSQL(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	_, err := repo.Claim(context.Background(), 100, now)
	if err != nil {
		t.Fatalf("Claim() unexpected error: %v", err)
	}

	if !strings.Contains(stub.gotSelectSQL, "FOR UPDATE SKIP LOCKED") {
		t.Fatalf("sql missing FOR UPDATE SKIP LOCKED:\n%s", stub.gotSelectSQL)
	}
	if !strings.Contains(stub.gotSelectSQL, "emitted_at IS NULL") {
		t.Fatalf("sql missing emitted_at IS NULL:\n%s", stub.gotSelectSQL)
	}
	if !strings.Contains(stub.gotSelectSQL, "next_attempt_at <= ?") {
		t.Fatalf("sql missing next_attempt_at condition:\n%s", stub.gotSelectSQL)
	}
	// args: now, now, limit
	if len(stub.gotSelectArgs) != 3 {
		t.Fatalf("args len = %d, want 3", len(stub.gotSelectArgs))
	}
	if stub.gotSelectArgs[2] != 100 {
		t.Fatalf("limit arg = %v, want 100", stub.gotSelectArgs[2])
	}
}

func TestClaim_returnsRecords(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	stub := &stubQE{
		selectRecords: []TaskEventRecord{
			{
				ID:            "evt-1",
				UserID:        "u-1",
				TaskID:        "t-1",
				RequestID:     "r-1",
				EventType:     "created",
				OccurredAt:    now,
				SchemaVersion: 1,
				Payload:       json.RawMessage(`{}`),
			},
		},
	}
	repo := New(stub)

	recs, err := repo.Claim(context.Background(), 500, now)
	if err != nil {
		t.Fatalf("Claim() unexpected error: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("records len = %d, want 1", len(recs))
	}
	if recs[0].ID != "evt-1" {
		t.Fatalf("record id = %s, want evt-1", recs[0].ID)
	}
}

func TestClaim_emptyResult(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	recs, err := repo.Claim(context.Background(), 500, now)
	if err != nil {
		t.Fatalf("Claim() unexpected error: %v", err)
	}
	if len(recs) != 0 {
		t.Fatalf("records len = %d, want 0", len(recs))
	}
}

// --- SetLease tests ---

func TestSetLease_emptyIDs_noop(t *testing.T) {
	t.Parallel()

	stub := &stubQE{}
	repo := New(stub)

	err := repo.SetLease(context.Background(), nil, "worker-1", 60*time.Second, time.Now())
	if err != nil {
		t.Fatalf("SetLease() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 0 {
		t.Fatalf("ExecContext should not be called for empty ids")
	}
}

func TestSetLease_buildsCorrectSQL(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 2}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)

	err := repo.SetLease(context.Background(), []string{"id-1", "id-2"}, "worker-1", 60*time.Second, now)
	if err != nil {
		t.Fatalf("SetLease() unexpected error: %v", err)
	}
	if len(stub.execCalls) != 1 {
		t.Fatalf("ExecContext call count = %d, want 1", len(stub.execCalls))
	}
	if !strings.Contains(stub.execCalls[0].sql, "lease_owner") {
		t.Fatalf("sql missing lease_owner:\n%s", stub.execCalls[0].sql)
	}
	if !strings.Contains(stub.execCalls[0].sql, "claimed_at") {
		t.Fatalf("sql missing claimed_at:\n%s", stub.execCalls[0].sql)
	}
}

// --- ExtendLease tests ---

func TestExtendLease_returnsAffectedRows(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 2}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	currentLease := now.Add(60 * time.Second)

	affected, err := repo.ExtendLease(
		context.Background(),
		[]string{"id-1", "id-2"},
		"worker-1",
		currentLease,
		60*time.Second,
		now,
	)
	if err != nil {
		t.Fatalf("ExtendLease() unexpected error: %v", err)
	}
	if affected != 2 {
		t.Fatalf("affected = %d, want 2", affected)
	}
}

func TestExtendLease_CAS_condition(t *testing.T) {
	t.Parallel()

	stub := &stubQE{execAffected: 0}
	repo := New(stub)
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	currentLease := now.Add(60 * time.Second)

	affected, err := repo.ExtendLease(
		context.Background(),
		[]string{"id-1"},
		"worker-1",
		currentLease,
		60*time.Second,
		now,
	)
	if err != nil {
		t.Fatalf("ExtendLease() unexpected error: %v", err)
	}
	// CASで一致しなかった場合 RowsAffected=0
	if affected != 0 {
		t.Fatalf("affected = %d, want 0 (CAS mismatch)", affected)
	}
	if len(stub.execCalls) != 1 {
		t.Fatalf("ExecContext call count = %d, want 1", len(stub.execCalls))
	}
	if !strings.Contains(stub.execCalls[0].sql, "lease_until = ?") {
		t.Fatalf("sql missing CAS condition (lease_until = ?):\n%s", stub.execCalls[0].sql)
	}
}

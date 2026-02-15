package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/logger"
)

// --- mock repo ---

type mockRepo struct {
	mu sync.Mutex

	claimRecords []ClaimedEvent
	claimErr     error
	claimCount   int

	setLeaseErr   error
	setLeaseCount int

	extendAffected int64
	extendErr      error

	markEmittedErr   error
	markEmittedCount int
	markEmittedIDs   []string

	markRetryErr   error
	markRetryCount int

	moveToDLQErr   error
	moveToDLQCount int

	releaseCount int
}

func (m *mockRepo) Claim(ctx context.Context, limit int, now time.Time) ([]ClaimedEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.claimCount++
	// 2回目以降は空を返す（無限ループ防止）
	if m.claimCount > 1 {
		return nil, nil
	}
	return m.claimRecords, m.claimErr
}

func (m *mockRepo) SetLease(ctx context.Context, ids []string, leaseOwner string, leaseDuration time.Duration, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setLeaseCount++
	return m.setLeaseErr
}

func (m *mockRepo) ReleaseLease(ctx context.Context, ids []string, leaseOwner string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.releaseCount++
	return nil
}

func (m *mockRepo) ExtendLease(ctx context.Context, ids []string, leaseOwner string, currentLeaseUntil time.Time, leaseDuration time.Duration, now time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.extendAffected, m.extendErr
}

func (m *mockRepo) MarkEmitted(ctx context.Context, ids []string, leaseOwner string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markEmittedCount++
	m.markEmittedIDs = ids
	return m.markEmittedErr
}

func (m *mockRepo) MarkRetry(ctx context.Context, ids []string, leaseOwner string, nextAttemptAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.markRetryCount++
	return m.markRetryErr
}

func (m *mockRepo) MoveToDLQ(ctx context.Context, ids []string, lastError string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.moveToDLQCount++
	return m.moveToDLQErr
}

// --- mock uploader ---

type mockUploader struct {
	mu          sync.Mutex
	uploadCount int
	uploadErr   error
	uploadKeys  []string

	existsResult bool
	existsErr    error
}

func (m *mockUploader) Upload(ctx context.Context, key string, body io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uploadCount++
	m.uploadKeys = append(m.uploadKeys, key)
	return m.uploadErr
}

func (m *mockUploader) Exists(ctx context.Context, key string) (bool, error) {
	return m.existsResult, m.existsErr
}

// --- helper ---

func testConfig() Config {
	cfg := DefaultConfig()
	cfg.IdleSleep = 10 * time.Millisecond // テスト高速化
	cfg.HeartbeatInterval = 1 * time.Hour  // テスト中は heartbeat 不要
	return cfg
}

type nopLogger struct{}

func (nopLogger) Debug(context.Context, string, error, ...logger.Attr) {}
func (nopLogger) Info(context.Context, string, ...logger.Attr)         {}
func (nopLogger) Warn(context.Context, string, ...logger.Attr)         {}
func (nopLogger) Error(context.Context, string, error, ...logger.Attr) {}

func testLogger() logger.Logger {
	return nopLogger{}
}

func mustRecords(ids ...string) []ClaimedEvent {
	now := time.Date(2026, 2, 1, 12, 0, 0, 0, time.UTC)
	recs := make([]ClaimedEvent, len(ids))
	for i, id := range ids {
		recs[i] = ClaimedEvent{
			ID:            id,
			UserID:        "u-1",
			TaskID:        "t-1",
			RequestID:     "r-1",
			EventType:     "created",
			OccurredAt:    now,
			SchemaVersion: 1,
			Payload:       json.RawMessage(`{}`),
			AttemptCount:  0,
		}
	}
	return recs
}

// --- tests ---

func TestProcessOnce_noEvents_returnsFalse(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{}
	up := &mockUploader{}
	w := NewWorker(repo, up, testConfig(), testLogger())

	processed, err := w.processOnce(context.Background())
	if err != nil {
		t.Fatalf("processOnce() error: %v", err)
	}
	if processed {
		t.Fatal("processOnce() should return false when no events")
	}
}

func TestProcessOnce_emitsEvents_markEmitted(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		claimRecords:   mustRecords("e1", "e2"),
		extendAffected: 2,
	}
	up := &mockUploader{}
	w := NewWorker(repo, up, testConfig(), testLogger())

	processed, err := w.processOnce(context.Background())
	if err != nil {
		t.Fatalf("processOnce() error: %v", err)
	}
	if !processed {
		t.Fatal("processOnce() should return true when events processed")
	}

	// S3 に 2回アップロード (data + manifest)
	if up.uploadCount != 2 {
		t.Fatalf("upload count = %d, want 2", up.uploadCount)
	}

	// MarkEmitted が呼ばれた
	if repo.markEmittedCount != 1 {
		t.Fatalf("markEmitted count = %d, want 1", repo.markEmittedCount)
	}
	if len(repo.markEmittedIDs) != 2 {
		t.Fatalf("markEmitted ids len = %d, want 2", len(repo.markEmittedIDs))
	}
}

func TestProcessOnce_s3Failure_markRetry(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		claimRecords:   mustRecords("e1"),
		extendAffected: 1,
	}
	up := &mockUploader{
		uploadErr: errors.New("s3 connection refused"),
	}
	w := NewWorker(repo, up, testConfig(), testLogger())

	processed, err := w.processOnce(context.Background())
	if err != nil {
		t.Fatalf("processOnce() error: %v", err)
	}
	if !processed {
		t.Fatal("processOnce() should return true even on failure")
	}

	// MarkRetry が呼ばれた
	if repo.markRetryCount != 1 {
		t.Fatalf("markRetry count = %d, want 1", repo.markRetryCount)
	}
	// MarkEmitted は呼ばれない
	if repo.markEmittedCount != 0 {
		t.Fatalf("markEmitted count = %d, want 0", repo.markEmittedCount)
	}
}

func TestProcessOnce_maxAttempt_movesToDLQ(t *testing.T) {
	t.Parallel()

	recs := mustRecords("e1")
	recs[0].AttemptCount = 4 // 次が5回目 → MaxAttempt(5) に到達

	repo := &mockRepo{
		claimRecords:   recs,
		extendAffected: 1,
	}
	up := &mockUploader{
		uploadErr: errors.New("s3 timeout"),
	}
	cfg := testConfig()
	cfg.MaxAttempt = 5
	w := NewWorker(repo, up, cfg, testLogger())

	processed, err := w.processOnce(context.Background())
	if err != nil {
		t.Fatalf("processOnce() error: %v", err)
	}
	if !processed {
		t.Fatal("processOnce() should return true")
	}

	// DLQ に移動
	if repo.moveToDLQCount != 1 {
		t.Fatalf("moveToDLQ count = %d, want 1", repo.moveToDLQCount)
	}
	// MarkRetry は呼ばれない
	if repo.markRetryCount != 0 {
		t.Fatalf("markRetry count = %d, want 0", repo.markRetryCount)
	}
}

func TestProcessOnce_manifestExists_skipsUpload(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{
		claimRecords:   mustRecords("e1"),
		extendAffected: 1,
	}
	up := &mockUploader{
		existsResult: true, // manifest が既に存在
	}
	w := NewWorker(repo, up, testConfig(), testLogger())

	processed, err := w.processOnce(context.Background())
	if err != nil {
		t.Fatalf("processOnce() error: %v", err)
	}
	if !processed {
		t.Fatal("processOnce() should return true")
	}

	// Upload は呼ばれない（manifest 存在 → スキップ）
	if up.uploadCount != 0 {
		t.Fatalf("upload count = %d, want 0 (manifest exists)", up.uploadCount)
	}
	// MarkEmitted は呼ばれる（DB 更新だけ再実行）
	if repo.markEmittedCount != 1 {
		t.Fatalf("markEmitted count = %d, want 1", repo.markEmittedCount)
	}
}

func TestRun_stopsOnContextCancel(t *testing.T) {
	t.Parallel()

	repo := &mockRepo{} // 常に0件
	up := &mockUploader{}
	cfg := testConfig()
	cfg.IdleSleep = 50 * time.Millisecond

	w := NewWorker(repo, up, cfg, testLogger())

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := w.Run(ctx)
	if err != nil {
		t.Fatalf("Run() should return nil on context cancel, got: %v", err)
	}
}

package compaction

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
)

// --- mock storage ---

type mockStorage struct {
	mu      sync.Mutex
	objects map[string][]byte
	uploads []string
}

func newMockStorage() *mockStorage {
	return &mockStorage{objects: make(map[string][]byte)}
}

func (m *mockStorage) Upload(ctx context.Context, key string, body io.Reader) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, _ := io.ReadAll(body)
	m.objects[key] = data
	m.uploads = append(m.uploads, key)
	return nil
}

func (m *mockStorage) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.objects[key]
	return ok, nil
}

func (m *mockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var keys []string
	for k := range m.objects {
		if strings.HasPrefix(k, prefix) {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

func (m *mockStorage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	data, ok := m.objects[key]
	if !ok {
		return nil, fmt.Errorf("not found: %s", key)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func putJSONL(m *mockStorage, key string, events []Event) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i := range events {
		if err := enc.Encode(events[i]); err != nil {
			panic(fmt.Sprintf("encode event: %v", err))
		}
	}
	m.objects[key] = buf.Bytes()
}

func testCompactionConfig() Config {
	return Config{
		BackfillWindow:  2,
		RawPrefix:       "raw/task-events",
		CompactedPrefix: "compacted/task-events",
		StatePrefix:     "_state/compaction",
	}
}

func testCompactionLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// --- tests ---

func TestWorker_Run_doneMarkerExists_skips(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	// done マーカーを事前配置
	storage.objects["_state/compaction/2026-02-11.done"] = []byte("done")

	w := NewWorker(storage, testCompactionConfig(), testCompactionLogger(), metrics.NewCompactionMetrics(metrics.NewProvider().Registry))
	target := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)

	err := w.Run(context.Background(), target)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// アップロードは発生しない
	if len(storage.uploads) != 0 {
		t.Fatalf("uploads = %d, want 0 (skipped)", len(storage.uploads))
	}
}

func TestWorker_Run_noEvents_writesDoneOnly(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	w := NewWorker(storage, testCompactionConfig(), testCompactionLogger(), metrics.NewCompactionMetrics(metrics.NewProvider().Registry))
	target := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)

	err := w.Run(context.Background(), target)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// done マーカーだけ書かれる
	if _, ok := storage.objects["_state/compaction/2026-02-11.done"]; !ok {
		t.Fatal("done marker not written")
	}
}

func TestWorker_Run_compactsAndWritesDone(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()

	// raw .jsonl を配置（target=2026-02-12, backfill=2 → 2026-02-12 と 2026-02-11 を読む）
	events := []Event{
		mustEvent("e1", "2026-02-11"),
		mustEvent("e2", "2026-02-11"),
		mustEvent("e3", "2026-02-12"),
	}
	putJSONL(storage, "raw/task-events/year=2026/month=02/day=12/hour=14/batch1.jsonl", events)

	w := NewWorker(storage, testCompactionConfig(), testCompactionLogger(), metrics.NewCompactionMetrics(metrics.NewProvider().Registry))
	target := time.Date(2026, 2, 12, 0, 0, 0, 0, time.UTC)

	err := w.Run(context.Background(), target)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// done マーカー
	if _, ok := storage.objects["_state/compaction/2026-02-12.done"]; !ok {
		t.Fatal("done marker not written")
	}

	// compaction manifest
	if _, ok := storage.objects["_state/compaction/2026-02-12.manifest.json"]; !ok {
		t.Fatal("compaction manifest not written")
	}

	// parquet ファイル（2日分のパーティション）
	hasDay11 := false
	hasDay12 := false
	for key := range storage.objects {
		if strings.Contains(key, "compacted/task-events/day=2026-02-11/") {
			hasDay11 = true
		}
		if strings.Contains(key, "compacted/task-events/day=2026-02-12/") {
			hasDay12 = true
		}
	}
	if !hasDay11 {
		t.Fatal("missing parquet for day=2026-02-11")
	}
	if !hasDay12 {
		t.Fatal("missing parquet for day=2026-02-12")
	}
}

func TestWorker_Run_deduplicatesEvents(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()

	// 重複を含む events
	events := []Event{
		mustEvent("e1", "2026-02-11"),
		mustEvent("e1", "2026-02-11"), // 重複
		mustEvent("e2", "2026-02-11"),
	}
	putJSONL(storage, "raw/task-events/year=2026/month=02/day=11/hour=10/batch1.jsonl", events)

	w := NewWorker(storage, testCompactionConfig(), testCompactionLogger(), metrics.NewCompactionMetrics(metrics.NewProvider().Registry))
	target := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)

	err := w.Run(context.Background(), target)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// manifest の内容を確認
	mData := storage.objects["_state/compaction/2026-02-11.manifest.json"]
	var m compactionManifest
	if err := json.Unmarshal(mData, &m); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}
	if m.TotalEvents != 2 {
		t.Fatalf("manifest total_events = %d, want 2", m.TotalEvents)
	}
	if m.DedupeRemoved != 1 {
		t.Fatalf("manifest dedupe_removed = %d, want 1", m.DedupeRemoved)
	}
}

func TestWorker_Run_manifestExists_writesDoneOnly(t *testing.T) {
	t.Parallel()

	storage := newMockStorage()
	// compaction manifest が既に存在（前回途中死亡ケース）
	storage.objects["_state/compaction/2026-02-11.manifest.json"] = []byte(`{}`)

	w := NewWorker(storage, testCompactionConfig(), testCompactionLogger(), metrics.NewCompactionMetrics(metrics.NewProvider().Registry))
	target := time.Date(2026, 2, 11, 0, 0, 0, 0, time.UTC)

	err := w.Run(context.Background(), target)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	// done マーカーだけ書かれる（parquet アップロードは不要）
	if _, ok := storage.objects["_state/compaction/2026-02-11.done"]; !ok {
		t.Fatal("done marker not written")
	}
	// uploads は done マーカーの1回だけ
	if len(storage.uploads) != 1 {
		t.Fatalf("uploads = %d, want 1 (done marker only)", len(storage.uploads))
	}
}

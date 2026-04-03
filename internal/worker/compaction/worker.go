package compaction

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
)

type Worker struct {
	storage ObjectStorage
	cfg     Config
	logger  *slog.Logger
	metrics *metrics.CompactionMetrics
}

func NewWorker(storage ObjectStorage, cfg Config, logger *slog.Logger, m *metrics.CompactionMetrics) *Worker {
	return &Worker{
		storage: storage,
		cfg:     cfg,
		logger:  logger,
		metrics: m,
	}
}

type compactionManifest struct {
	ClaimDate      string             `json:"claim_date"`
	OutputFiles    []compactionOutput `json:"output_files"`
	InputManifests int                `json:"input_manifests"`
	TotalEvents    int                `json:"total_events"`
	DedupeRemoved  int                `json:"dedupe_removed"`
	CreatedAt      time.Time          `json:"created_at"`
}

type compactionOutput struct {
	Day   string `json:"day"`
	Key   string `json:"key"`
	Count int    `json:"count"`
}

func (w *Worker) Run(ctx context.Context, target time.Time) error {
	claimDate := target.Format("2006-01-02")
	w.logger.Info("compaction started", "claim_date", claimDate)

	doneKey := fmt.Sprintf("%s/%s.done", w.cfg.StatePrefix, claimDate)
	done, err := w.storage.Exists(ctx, doneKey)
	if err != nil {
		return fmt.Errorf("check done marker: %w", err)
	}
	if done {
		w.logger.Info("already compacted, skipping", "claim_date", claimDate)
		return nil
	}

	manifestKey := fmt.Sprintf("%s/%s.manifest.json", w.cfg.StatePrefix, claimDate)
	manifestExists, err := w.storage.Exists(ctx, manifestKey)
	if err != nil {
		return fmt.Errorf("check compaction manifest: %w", err)
	}
	if manifestExists {

		w.logger.Info("compaction manifest exists, writing done marker only")
		return w.writeDoneMarker(ctx, doneKey)
	}

	claimDates := claimDatesForTarget(target, w.cfg.BackfillWindow)

	events, err := readRawEvents(ctx, w.storage, w.cfg.RawPrefix, claimDates, w.metrics)
	if err != nil {
		return fmt.Errorf("read raw events: %w", err)
	}
	w.metrics.EventsRead.Set(float64(len(events)))
	if len(events) == 0 {
		w.logger.Info("no events found, skipping", "claim_date", claimDate)
		return w.writeDoneMarker(ctx, doneKey)
	}

	deduped, removed := dedupe(events)
	w.metrics.EventsDeduped.Set(float64(len(deduped)))
	w.logger.Info("dedupe done", "total", len(events), "deduped", len(deduped), "removed", removed)

	groups := groupByDay(deduped)

	var outputs []compactionOutput
	for day, dayEvents := range groups {
		if err := uploadParquet(ctx, w.storage, w.cfg.CompactedPrefix, day, claimDate, dayEvents, w.metrics); err != nil {
			return fmt.Errorf("upload parquet day=%s: %w", day, err)
		}
		outputs = append(outputs, compactionOutput{
			Day:   day,
			Key:   compactedKey(w.cfg.CompactedPrefix, day, claimDate),
			Count: len(dayEvents),
		})
	}
	w.metrics.ParquetFiles.Set(float64(len(outputs)))

	m := compactionManifest{
		ClaimDate:      claimDate,
		OutputFiles:    outputs,
		InputManifests: 0,
		TotalEvents:    len(deduped),
		DedupeRemoved:  removed,
		CreatedAt:      time.Now(),
	}
	if err := w.writeManifest(ctx, manifestKey, m); err != nil {
		return fmt.Errorf("write compaction manifest: %w", err)
	}

	if err := w.writeDoneMarker(ctx, doneKey); err != nil {
		return err
	}

	w.logger.Info("compaction completed",
		"claim_date", claimDate,
		"events", len(deduped),
		"days", len(groups),
		"removed", removed,
	)
	return nil
}

func (w *Worker) writeManifest(ctx context.Context, key string, m compactionManifest) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("encode compaction manifest: %w", err)
	}
	return w.storage.Upload(ctx, key, bytes.NewReader(buf.Bytes()))
}

func (w *Worker) writeDoneMarker(ctx context.Context, key string) error {
	return w.storage.Upload(ctx, key, bytes.NewReader([]byte("done")))
}

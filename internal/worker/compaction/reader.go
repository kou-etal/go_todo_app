package compaction

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
)

func readRawEvents(
	ctx context.Context,
	storage ObjectStorage,
	rawPrefix string,

	claimDates []time.Time,
	metrics *metrics.CompactionMetrics,
) ([]Event, error) {
	var allEvents []Event

	for _, date := range claimDates {
		events, err := readEventsForDate(ctx, storage, rawPrefix, date, metrics)
		if err != nil {
			return nil, fmt.Errorf("read events for %s: %w", date.Format("2006-01-02"), err)
		}
		allEvents = append(allEvents, events...)
	}
	return allEvents, nil
}

func readEventsForDate(
	ctx context.Context,
	storage ObjectStorage,
	rawPrefix string,
	date time.Time,
	metrics *metrics.CompactionMetrics,
) ([]Event, error) {
	var events []Event

	for hour := 0; hour < 24; hour++ {
		prefix := fmt.Sprintf("%s/year=%04d/month=%02d/day=%02d/hour=%02d/",
			rawPrefix, date.Year(), date.Month(), date.Day(), hour,
		)

		listStart := time.Now()
		keys, err := storage.List(ctx, prefix)
		listDur := time.Since(listStart).Seconds()
		metrics.S3ListDuration.Add(listDur)
		metrics.S3ListCalls.Add(1)
		metrics.S3ObjectsListed.Add(float64(len(keys)))
		if err != nil {
			return nil, fmt.Errorf("list prefix=%s: %w", prefix, err)
		}

		for _, key := range keys {
			if !strings.HasSuffix(key, ".jsonl") {
				continue
			}
			readStart := time.Now()
			parsed, err := readJSONLFile(ctx, storage, key)
			metrics.S3ReadDuration.Add(time.Since(readStart).Seconds())
			if err != nil {
				return nil, fmt.Errorf("read jsonl key=%s: %w", key, err)
			}
			events = append(events, parsed...)
		}
	}
	return events, nil
}

func readJSONLFile(ctx context.Context, storage ObjectStorage, key string) ([]Event, error) {
	body, err := storage.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := body.Close(); err != nil {
			slog.Error("failed to close S3 body", slog.String("error", err.Error()))
		}
	}()

	var events []Event
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			return nil, fmt.Errorf("unmarshal line: %w", err)
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan jsonl: %w", err)
	}
	return events, nil
}

func claimDatesForTarget(target time.Time, backfillWindow int) []time.Time {
	dates := make([]time.Time, backfillWindow)
	for i := range backfillWindow {

		dates[i] = target.AddDate(0, 0, -i)
	}
	return dates
}

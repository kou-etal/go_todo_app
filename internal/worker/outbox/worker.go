package outbox

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
)

type Worker struct {
	repo     OutboxRepo
	uploader ObjectUploader
	cfg      Config
	ownerID  string
	s3Prefix string
	logger   logger.Logger
	metrics  *metrics.OutboxMetrics
}

func NewWorker(repo OutboxRepo, uploader ObjectUploader, cfg Config, logger logger.Logger, m *metrics.OutboxMetrics) *Worker {
	return &Worker{
		repo:     repo,
		uploader: uploader,
		cfg:      cfg,
		ownerID:  uuid.New().String(),
		s3Prefix: normalizePrefix("task-events"),
		logger:   logger,
		metrics:  m,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info(ctx, "outbox worker started", logger.String("owner", w.ownerID))

	go w.queueDepthLoop(ctx)

	for {

		select {
		case <-ctx.Done():
			w.logger.Info(ctx, "outbox worker shutting down")
			return nil
		default:
		}

		processed, err := w.processOnce(ctx)
		if err != nil {
			w.logger.Error(ctx, "outbox worker process error", err)
		}

		if !processed {
			w.metrics.IdleCycles.Inc()
			select {

			case <-ctx.Done():
				return nil
			case <-time.After(w.cfg.IdleSleep):
			}
		}
	}
}

func (w *Worker) processOnce(ctx context.Context) (bool, error) {
	processStart := time.Now()
	defer func() {
		w.metrics.ProcessDuration.Observe(time.Since(processStart).Seconds())
	}()

	now := time.Now()

	claimStart := time.Now()
	records, err := w.repo.Claim(ctx, w.cfg.ChunkMaxRows, now)
	w.metrics.ClaimDuration.Observe(time.Since(claimStart).Seconds())
	if err != nil {
		w.metrics.RepoFailures.WithLabelValues("claim").Inc()
		return false, fmt.Errorf("claim: %w", err)
	}

	w.metrics.ClaimBatchSize.Observe(float64(len(records)))
	if len(records) == 0 {
		return false, nil
	}

	for _, r := range records {
		w.metrics.EventLagToClaim.Observe(now.Sub(r.OccurredAt).Seconds())
	}

	records = trimByByteLimit(records, w.cfg.ChunkMaxBytes)

	if len(records) == 0 {
		return false, nil
	}

	ids := collectIDs(records)

	now = time.Now()

	if err := w.repo.SetLease(ctx, ids, w.ownerID, w.cfg.LeaseDuration, now); err != nil {
		w.metrics.RepoFailures.WithLabelValues("set_lease").Inc()
		return false, fmt.Errorf("set lease: %w", err)
	}
	leaseUntil := now.Add(w.cfg.LeaseDuration)
	claimedAt := now

	w.metrics.InflightLeased.Set(float64(len(ids)))
	defer w.metrics.InflightLeased.Set(0)

	hbCtx, hbCancel := context.WithCancel(ctx)
	defer hbCancel()
	go w.heartbeatLoop(hbCtx, ids, leaseUntil)

	emitErr := w.emitToS3(ctx, records, ids, claimedAt)
	hbCancel()

	if emitErr != nil {
		w.handleFailure(ctx, ids, records, emitErr)
		return true, nil
	}

	emitNow := time.Now()
	if err := w.repo.MarkEmitted(ctx, ids, w.ownerID, emitNow); err != nil {
		w.metrics.RepoFailures.WithLabelValues("mark_emitted").Inc()
		return true, fmt.Errorf("mark emitted: %w", err)
	}

	for _, r := range records {
		w.metrics.EventLagToEmit.Observe(emitNow.Sub(r.OccurredAt).Seconds())
	}

	w.metrics.EventsEmitted.Add(float64(len(ids)))
	w.logger.Info(ctx, "emitted events", logger.Int("count", len(ids)))
	return true, nil
}

func (w *Worker) emitToS3(
	ctx context.Context,
	records []ClaimedEvent,
	ids []string,
	claimedAt time.Time,
) error {

	bid := batchID(ids, claimedAt, 1)
	dataKey := s3DataKey(w.s3Prefix, claimedAt, bid)
	manifestKey := s3ManifestKey(w.s3Prefix, claimedAt, bid)

	exists, err := w.uploader.Exists(ctx, manifestKey)
	if err != nil {
		return fmt.Errorf("check manifest exists: %w", err)
	}
	if exists {
		w.logger.Info(ctx, "manifest already exists, skipping upload", logger.String("key", manifestKey))

		return nil
	}

	jsonl, err := buildJSONLines(records)

	if err != nil {
		return fmt.Errorf("build jsonl: %w", err)
	}
	dataStart := time.Now()
	if err := w.uploader.Upload(ctx, dataKey, bytes.NewReader(jsonl)); err != nil {

		return fmt.Errorf("upload data: %w", err)
	}
	w.metrics.UploadDataDuration.Observe(time.Since(dataStart).Seconds())

	m := manifest{
		BatchID:   bid,
		DataKey:   dataKey,
		EventIDs:  ids,
		Count:     len(ids),
		CreatedAt: time.Now(),
	}
	mJSON, err := buildManifestJSON(m)
	if err != nil {
		return fmt.Errorf("build manifest json: %w", err)
	}
	manifestStart := time.Now()
	if err := w.uploader.Upload(ctx, manifestKey, bytes.NewReader(mJSON)); err != nil {
		return fmt.Errorf("upload manifest: %w", err)
	}
	w.metrics.UploadManifestDuration.Observe(time.Since(manifestStart).Seconds())

	return nil

}

func (w *Worker) handleFailure(
	ctx context.Context,
	ids []string,
	records []ClaimedEvent,
	emitErr error,
) {
	w.logger.Error(ctx, "emit failed", emitErr, logger.Int("count", len(ids)))

	var dlqIDs []string
	var retryIDs []string
	var maxRetryCount uint32

	for _, r := range records {
		if r.AttemptCount+1 >= w.cfg.MaxAttempt {
			dlqIDs = append(dlqIDs, r.ID)
		} else {
			retryIDs = append(retryIDs, r.ID)
			if r.AttemptCount >= maxRetryCount {
				maxRetryCount = r.AttemptCount
			}
		}
	}

	if len(dlqIDs) > 0 {
		w.metrics.EventsDLQ.Add(float64(len(dlqIDs)))
		w.logger.Warn(ctx, "moving to DLQ", logger.Attr{Key: "ids", Value: dlqIDs}, logger.Int("count", len(dlqIDs)))
		if err := w.repo.MoveToDLQ(ctx, dlqIDs, emitErr.Error(), time.Now()); err != nil {
			w.metrics.RepoFailures.WithLabelValues("move_dlq").Inc()

			w.logger.Error(ctx, "move to DLQ failed", err)
		}
	}

	if len(retryIDs) > 0 {
		w.metrics.EventsRetried.Add(float64(len(retryIDs)))
		nextAt := NextAttemptAt(time.Now(), int(maxRetryCount)+1, w.cfg.BackoffBase)
		if err := w.repo.MarkRetry(ctx, retryIDs, w.ownerID, nextAt); err != nil {
			w.metrics.RepoFailures.WithLabelValues("mark_retry").Inc()
			w.logger.Error(ctx, "mark retry failed", err)
		}
	}
}

func (w *Worker) heartbeatLoop(ctx context.Context, ids []string, currentLeaseUntil time.Time) {
	ticker := time.NewTicker(w.cfg.HeartbeatInterval)

	defer ticker.Stop()

	for {

		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			now := time.Now()
			extendStart := time.Now()
			affected, err := w.repo.ExtendLease(
				ctx, ids, w.ownerID,
				currentLeaseUntil, w.cfg.LeaseDuration, now,
			)
			w.metrics.HeartbeatExtendDuration.Observe(time.Since(extendStart).Seconds())
			if err != nil {
				w.metrics.RepoFailures.WithLabelValues("extend_lease").Inc()
				w.logger.Error(ctx, "heartbeat extend lease error", err)
				return
			}
			w.metrics.HeartbeatExtendAffected.Observe(float64(affected))
			if affected == 0 {
				w.metrics.HeartbeatLost.Inc()
				w.logger.Warn(ctx, "heartbeat CAS mismatch, lease lost")
				return
			}
			currentLeaseUntil = now.Add(w.cfg.LeaseDuration)

		}
	}
}

func (w *Worker) queueDepthLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := w.repo.CountUnemitted(ctx)
			if err != nil {
				w.logger.Error(ctx, "queue depth count error", err)
				continue
			}
			w.metrics.QueueDepth.Set(float64(count))
		}
	}
}

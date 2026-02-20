package outbox

import "github.com/prometheus/client_golang/prometheus"

var (
	outboxProcessDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_process_duration_seconds",
		Help: "processOnce 1サイクルの所要時間",
	})
	outboxClaimDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_claim_duration_seconds",
		Help: "Claim (SELECT FOR UPDATE) の所要時間",
	})
	outboxClaimBatchSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "outbox_claim_batch_size",
		Help:    "Claim で取得したレコード数",
		Buckets: []float64{0, 1, 10, 50, 100, 250, 500},
	})
	outboxUploadDataDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_upload_data_duration_seconds",
		Help: "data (.jsonl) の S3 PUT 所要時間",
	})
	outboxUploadManifestDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_upload_manifest_duration_seconds",
		Help: "manifest の S3 PUT 所要時間",
	})
	outboxHeartbeatExtendDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_heartbeat_extend_duration_seconds",
		Help: "ExtendLease 1回の所要時間",
	})
	outboxHeartbeatExtendAffected = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "outbox_heartbeat_extend_affected_rows",
		Help:    "ExtendLease の affected rows",
		Buckets: []float64{0, 1, 10, 50, 100, 250, 500},
	})
	outboxEventLagToClaim = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_event_lag_to_claim_seconds",
		Help: "occurred_at から claim 開始までの遅延",
	})
	outboxEventLagToEmit = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "outbox_event_lag_to_emit_seconds",
		Help: "occurred_at から emit 完了までの遅延",
	})

	// Counter

	outboxEventsEmitted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_events_emitted_total",
		Help: "emit 成功したイベント数",
	})
	outboxEventsRetried = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_events_retried_total",
		Help: "retry に回ったイベント数",
	})
	outboxEventsDLQ = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_events_dlq_total",
		Help: "DLQ に移動したイベント数",
	})
	outboxRepoFailures = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "outbox_repo_failures_total",
		Help: "DB 操作失敗数",
	}, []string{"op"})
	outboxHeartbeatLost = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_heartbeat_lost_total",
		Help: "CAS mismatch でリース喪失した回数",
	})
	outboxIdleCycles = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "outbox_idle_cycles_total",
		Help: "Claim 0件で sleep に入った回数",
	})

	// Gauge

	outboxInflightLeased = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_inflight_leased",
		Help: "現在 lease 中の件数",
	})
	outboxQueueDepth = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_queue_depth",
		Help: "未 emit のイベント数（低頻度計測）",
	})
)

func init() {
	prometheus.MustRegister(
		outboxProcessDuration,
		outboxClaimDuration,
		outboxClaimBatchSize,
		outboxUploadDataDuration,
		outboxUploadManifestDuration,
		outboxHeartbeatExtendDuration,
		outboxHeartbeatExtendAffected,
		outboxEventLagToClaim,
		outboxEventLagToEmit,
		outboxEventsEmitted,
		outboxEventsRetried,
		outboxEventsDLQ,
		outboxRepoFailures,
		outboxHeartbeatLost,
		outboxIdleCycles,
		outboxInflightLeased,
		outboxQueueDepth,
	)
}

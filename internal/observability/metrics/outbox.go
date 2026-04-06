package metrics

import "github.com/prometheus/client_golang/prometheus"

type OutboxMetrics struct {
	ProcessDuration         prometheus.Histogram
	ClaimDuration           prometheus.Histogram
	ClaimBatchSize          prometheus.Histogram
	UploadDataDuration      prometheus.Histogram
	UploadManifestDuration  prometheus.Histogram
	HeartbeatExtendDuration prometheus.Histogram
	HeartbeatExtendAffected prometheus.Histogram
	EventLagToClaim         prometheus.Histogram
	EventLagToEmit          prometheus.Histogram

	EventsEmitted prometheus.Counter
	EventsRetried prometheus.Counter
	EventsDLQ     prometheus.Counter
	RepoFailures  *prometheus.CounterVec
	HeartbeatLost prometheus.Counter
	IdleCycles    prometheus.Counter

	InflightLeased prometheus.Gauge
	QueueDepth     prometheus.Gauge
}

//ここで分岐。
func NewOutboxMetrics(reg *prometheus.Registry) *OutboxMetrics {
	m := &OutboxMetrics{
		ProcessDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_process_duration_seconds",
			Help: "processOnce 1サイクルの所要時間",
		}),
		ClaimDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_claim_duration_seconds",
			Help: "Claim (SELECT FOR UPDATE) の所要時間",
		}),
		ClaimBatchSize: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "outbox_claim_batch_size",
			Help:    "Claim で取得したレコード数",
			Buckets: []float64{0, 1, 10, 50, 100, 250, 500},
		}),
		UploadDataDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_upload_data_duration_seconds",
			Help: "data (.jsonl) の S3 PUT 所要時間",
		}),
		UploadManifestDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_upload_manifest_duration_seconds",
			Help: "manifest の S3 PUT 所要時間",
		}),
		HeartbeatExtendDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_heartbeat_extend_duration_seconds",
			Help: "ExtendLease 1回の所要時間",
		}),
		HeartbeatExtendAffected: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "outbox_heartbeat_extend_affected_rows",
			Help:    "ExtendLease の affected rows",
			Buckets: []float64{0, 1, 10, 50, 100, 250, 500},
		}),
		EventLagToClaim: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_event_lag_to_claim_seconds",
			Help: "occurred_at から claim 開始までの遅延",
		}),
		EventLagToEmit: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name: "outbox_event_lag_to_emit_seconds",
			Help: "occurred_at から emit 完了までの遅延",
		}),

		EventsEmitted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_emitted_total",
			Help: "emit 成功したイベント数",
		}),
		EventsRetried: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_retried_total",
			Help: "retry に回ったイベント数",
		}),
		EventsDLQ: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_dlq_total",
			Help: "DLQ に移動したイベント数",
		}),
		RepoFailures: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "outbox_repo_failures_total",
			Help: "DB 操作失敗数",
		}, []string{"op"}),
		HeartbeatLost: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_heartbeat_lost_total",
			Help: "CAS mismatch でリース喪失した回数",
		}),
		IdleCycles: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_idle_cycles_total",
			Help: "Claim 0件で sleep に入った回数",
		}),

		InflightLeased: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outbox_inflight_leased",
			Help: "現在 lease 中の件数",
		}),
		QueueDepth: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "outbox_queue_depth",
			Help: "未 emit のイベント数（低頻度計測）",
		}),
	}

	reg.MustRegister( //登録
		m.ProcessDuration,
		m.ClaimDuration,
		m.ClaimBatchSize,
		m.UploadDataDuration,
		m.UploadManifestDuration,
		m.HeartbeatExtendDuration,
		m.HeartbeatExtendAffected,
		m.EventLagToClaim,
		m.EventLagToEmit,
		m.EventsEmitted,
		m.EventsRetried,
		m.EventsDLQ,
		m.RepoFailures,
		m.HeartbeatLost,
		m.IdleCycles,
		m.InflightLeased,
		m.QueueDepth,
	)

	return m
}

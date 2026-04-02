package metrics

import "github.com/prometheus/client_golang/prometheus"

type CompactionMetrics struct {
	LastSuccessTimestamp prometheus.Gauge
	LastFailureTimestamp prometheus.Gauge
	LastRunStatus        prometheus.Gauge
	LastRunDuration      prometheus.Gauge
	EventsRead          prometheus.Gauge
	EventsDeduped       prometheus.Gauge
	ParquetFiles        prometheus.Gauge
	S3ListDuration      prometheus.Gauge
	S3ListCalls         prometheus.Gauge
	S3ListPages         prometheus.Gauge
	S3ObjectsListed     prometheus.Gauge
	S3ReadDuration      prometheus.Gauge
	ParquetWriteDuration prometheus.Gauge
	S3UploadDuration    prometheus.Gauge
}

func NewCompactionMetrics(reg *prometheus.Registry) *CompactionMetrics {
	m := &CompactionMetrics{
		LastSuccessTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_last_success_timestamp",
			Help: "最後に成功した時刻 (unix epoch)",
		}),
		LastFailureTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_last_failure_timestamp",
			Help: "最後に失敗した時刻 (unix epoch)",
		}),
		LastRunStatus: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_last_run_status",
			Help: "直近の実行結果 (1=成功, 0=失敗)",
		}),
		LastRunDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_last_run_duration_seconds",
			Help: "直近の実行所要時間",
		}),
		EventsRead: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_events_read",
			Help: "読み込んだイベント数",
		}),
		EventsDeduped: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_events_deduped",
			Help: "dedupe 後のイベント数",
		}),
		ParquetFiles: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_parquet_files",
			Help: "生成した Parquet ファイル数",
		}),
		S3ListDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_list_duration_seconds",
			Help: "S3 List 操作の合計所要時間",
		}),
		S3ListCalls: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_list_calls",
			Help: "S3 List API コール数",
		}),
		S3ListPages: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_list_pages",
			Help: "S3 List ページネーション数",
		}),
		S3ObjectsListed: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_objects_listed",
			Help: "S3 List で返されたオブジェクト数",
		}),
		S3ReadDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_read_duration_seconds",
			Help: "S3 Get (jsonl 読み込み) の合計所要時間",
		}),
		ParquetWriteDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_parquet_write_duration_seconds",
			Help: "Parquet 変換の合計所要時間",
		}),
		S3UploadDuration: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "compaction_s3_upload_duration_seconds",
			Help: "S3 Upload (Parquet) の合計所要時間",
		}),
	}

	reg.MustRegister(
		m.LastSuccessTimestamp,
		m.LastFailureTimestamp,
		m.LastRunStatus,
		m.LastRunDuration,
		m.EventsRead,
		m.EventsDeduped,
		m.ParquetFiles,
		m.S3ListDuration,
		m.S3ListCalls,
		m.S3ListPages,
		m.S3ObjectsListed,
		m.S3ReadDuration,
		m.ParquetWriteDuration,
		m.S3UploadDuration,
	)

	return m
}

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RED メソッド。定番の三つの型。
// Requests Per Second=RPS。1秒あたり何リクエスト扱ってる。
// Vec = ラベル付きメトリクス。同じメトリクスをラベルで分割できる。
// requestsTotal = 150ではなくrequestsTotal{method="GET",  path="/tasks", status="200"} = 100
// Vec にしておけば「GET /tasks の 5xx だけ」みたいなフィルタが Grafana でできる。
type HTTPMetrics struct {
	RequestsTotal *prometheus.CounterVec // Rate + Errors
	//Duration = 1リクエストの作業にかかった時間。
	RequestDuration *prometheus.HistogramVec //Duration
	/*
			P50/P95/P99 = パーセンタイル。

		  100リクエストのレイテンシを順に並べた時：

		  P50 = 50番目の値  → 50%のリクエストはこれに収まってる
		  P95 = 95番目の値  → 95%はこれに収まってる。残り5%が遅い
		  P99 = 99番目の値  → ほぼ全部これに収まってる。最悪に近いケース
	*/
	RequestsInFlight prometheus.Gauge
	//InFlight は総数だけ見ればいい->vec使わないgauge
}

func NewHTTPMetrics(reg *prometheus.Registry) *HTTPMetrics {
	m := &HTTPMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "HTTP リクエスト数",
			},
			[]string{"method", "path", "status"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP リクエストのレイテンシ",
				Buckets: prometheus.DefBuckets,
				//Histogramの区切りの定義。デフォルトを使用。
				// /5msは何件、10msは何件、25msは何件...」と振り分ける。
			},
			[]string{"method", "path", "status"},
		),
		RequestsInFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "http_requests_in_flight",
				Help: "処理中の HTTP リクエスト数",
			},
		),
	}

	reg.MustRegister(m.RequestsTotal, m.RequestDuration, m.RequestsInFlight)
	return m
}

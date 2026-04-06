package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
)

// Metrics は HTTP メトリクスを計測するミドルウェアを返す。
func Metrics(m *metrics.HTTPMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m.RequestsInFlight.Inc()       //現在作業中の総数。
			defer m.RequestsInFlight.Dec() //パニックしても減るようにしてる。

			start := time.Now()
			rec := &statusRecorder{w: w, statusCode: http.StatusOK}

			next.ServeHTTP(rec, r)

			status := strconv.Itoa(rec.statusCode)
			path := normalizePath(r.URL.Path)
			dur := time.Since(start).Seconds()

			m.RequestsTotal.WithLabelValues(r.Method, path, status).Inc()          // 完了したら+1( 累計何件)
			m.RequestDuration.WithLabelValues(r.Method, path, status).Observe(dur) //完了したら経過時間を記録（どれくらいかかった）
		})
	}
}

// http.ResponseWriter は WriteHeader() を1回呼んだらステータスコードが確定して送信される。
// handlerの中で書き込まれた後からステータスコードを取得する方法がないから、statusRecorder で WriteHeaderを横取りする。

// TODO:コードがaccessmiddlewareと共通やからまとめる。
type statusRecorder struct {
	w           http.ResponseWriter
	statusCode  int
	wroteHeader bool
}

func (r *statusRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = statusCode
	r.wroteHeader = true
	r.w.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.w.Write(p)
}

func (r *statusRecorder) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

// normalizePath はパスを正規化してカーディナリティ爆発を防ぐ。
// /tasks/550e8400-e29b-... → /tasks/{id}
// 正規化しないと、リクエストごとに違うラベルが生まれる
// Prometheus はラベルの組み合わせごとに時系列データを保持するから、メモリもストレージも爆発する
func normalizePath(p string) string {
	if strings.HasPrefix(p, "/tasks/") {
		return "/tasks/{id}"
	}
	return p
}

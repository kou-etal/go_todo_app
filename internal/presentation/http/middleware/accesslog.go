package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kou-etal/go_todo_app/internal/logger"
)

type responseRecorder struct {
	w           http.ResponseWriter
	statusCode  int
	wroteHeader bool
	bytes       int64
}

func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = statusCode
	r.wroteHeader = true
	r.w.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.w.Write(p)
	r.bytes += int64(n)
	return n, err
}

func (r *responseRecorder) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

func AccessLog(lg logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				w:          w,
				statusCode: http.StatusOK,
			}

			next.ServeHTTP(rec, r)

			dur := time.Since(start)

			attrs := []logger.Attr{

				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", rec.statusCode),
				logger.Int64("bytes", rec.bytes),
				logger.Int64("duration_us", dur.Microseconds()),
				logger.String("ip", clientIP(r)),
				logger.String("ua", userAgent(r)),
			}

			if rec.statusCode >= 500 {
				lg.Error(r.Context(), "access", nil, attrs...)

				return

			}
			lg.Info(r.Context(), "access", attrs...)
		})
	}
}

// OS、ブラウザなどのデータ。これで特定のブラウザだけエラーとかbotとかが分かる。
func userAgent(r *http.Request) string {
	//長すぎるuseragentを弾く
	ua := r.UserAgent()
	if len(ua) > 200 {
		return ua[:200] //これ便利記法
	}
	return ua
}

// RemoteAddrは直接つないできたIP。client->CDN->LB->app
// X-Forwarded-ForはクライアントIPまで。cloudflare、aws alb、nginxとかがつけてくれる。
func clientIP(r *http.Request) string {
	//クライアントのIPを取る。r.RemoteAddrはプロキシのIP。203.0.113.1, 10.0.0.5など中継点のIPもとってきつつ
	//一番左、つまり真のクライアントIPを取る。
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",") //"203.0.113.1, 10.0.0.5"->["203.0.113.1", " 10.0.0.5"]
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	//XFF が無い環境用
	//X-Real-IPこれはhttpヘッダー。RemoteAddrはTCPコネクションの相手->Goが所有。
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	//XFFもReal-IPも無い場合->ローカル開発。LBやnginxを介してない単純構成。

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	//r.RemoteAddrはip:port
	//192.168.1.10:54321の左部分だけ。IPv6とかになると文字列処理ミスりやすいから標準のSplitHostPort使う。host, port, err

	return r.RemoteAddr //最後の保険
}

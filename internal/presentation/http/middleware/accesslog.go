package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kou-etal/go_todo_app/internal/logger"
)

// ResponseWriterをラップ
type responseRecorder struct {
	w           http.ResponseWriter
	statusCode  int
	wroteHeader bool
	bytes       int64
}

/*透過ってのは丸めてしまったメソッドの再生。
http.ResponseWriterはhttp.Flusher持ってるけど(interfaceがFlusherを持ってるわけじゃない)
それをラップするとresponserecoder型になりw.http.Flusherは不可。ゆえに復活させる、それが透過*/

// net/httpが標準でstatusを返さないから自前で観測している
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

// flush クライアントに即データ返す
// 今は別にflush使わんけどstreamingとかで使う。その時にpanic起こさないように保険
func (r *responseRecorder) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

//Hijack/Push/ReadFrom等が必要になったらここに透過実装

func AccessLog(lg logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				w:          w,
				statusCode: http.StatusOK, //デフォルトを定義
			}

			next.ServeHTTP(rec, r)
			//handler/router/usecaseがここで全部実行される
			//ここでさっき定義したwriteheader使う。

			dur := time.Since(start)

			attrs := []logger.Attr{
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path), // queryはログしない
				logger.Int("status", rec.statusCode),
				logger.Int("bytes", int(rec.bytes)),
				logger.Int("duration_ms", int(dur.Milliseconds())),
				logger.String("ip", clientIP(r)),
				logger.String("ua", userAgent(r)),
			} //比較、計算したい奴はintで返す。metrics、アラート条件で使う。

			if rec.statusCode >= 500 {
				lg.Error(r.Context(), "access", nil, attrs...)
				//エラー本体は返さない、
				return
				//ErrorはアラートなるからaccesslogはInfo固定にする場合もある
			}
			lg.Info(r.Context(), "access", attrs...)
		})
	}
}

func userAgent(r *http.Request) string {
	//長すぎるuseragentを弾く
	ua := r.UserAgent()
	if len(ua) > 200 {
		return ua[:200]
	}
	return ua
}

func clientIP(r *http.Request) string {
	//クライアントのIPを取る。r.RemoteAddrはプロキシのIP。203.0.113.1, 10.0.0.5など中継点のIPもとってきつつ
	//一番左、つまり真のクライアントIPを取る。
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	//XFF が無い環境用
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		return xrip
	}
	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil && host != "" {
		return host
	}
	//192.168.1.10:54321の左部分だけ

	return r.RemoteAddr //最後の保険
}

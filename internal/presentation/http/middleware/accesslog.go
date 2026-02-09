package middleware

//ここはそもそもmiddlewareやからhttp通信は全部通る。ドメイン単位の観測とは異なる。
import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kou-etal/go_todo_app/internal/logger"
)

// ResponseWriterをラップ
// ResponseWriterは最終的に返したstatus code/バイト数は受け取れない。アクセスログではそれが欲しい。だからラップする
type responseRecorder struct {
	w           http.ResponseWriter
	statusCode  int
	wroteHeader bool
	bytes       int64
}

/*透過ってのは丸めてしまったメソッドの再生。
http.ResponseWriterはhttp.Flusher持ってるけど(interfaceがFlusherを持ってるわけじゃない)
それをラップするとresponserecoder型になりw.http.Flusherは不可。ゆえに復活させる、それが透過*/

//interfaceはmin機能だけ要求してそれ以外の機能は持ってたらいいよね->能力ベース設計。差し替えが簡単になる。
//それがhijackとか

// net/httpが標準でstatusを返さないから自前で観測している
func (r *responseRecorder) Header() http.Header {
	return r.w.Header()
}

// net/httpのWriteHeaderは一回目の書き込みを信用し、二回目以降は捨てる。それに合わせた実装。
func (r *responseRecorder) WriteHeader(statusCode int) {
	if r.wroteHeader {
		return
	}
	r.statusCode = statusCode
	r.wroteHeader = true
	r.w.WriteHeader(statusCode)
}

// json.NewEncoder(w).Encode(v)は中でmarshal→w.Write(bytes)。fmt.Fprint(w, "hello")は中でw.Write(bytes)。
func (r *responseRecorder) Write(p []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	} //header無いのはミスだからエラーにすべきでは->ミスじゃないらしい。200をデフォルトにするのが標準。それに合わせる。
	// 別にheaderなくてもアプリは詰まない。
	n, err := r.w.Write(p) //nはライフサイクル短いしいい命名。lは1とかIで読みにくいらしい。
	r.bytes += int64(n)    //Writeは何回も使われる->環境次第ではint32になってオーバーする。
	// int64に寄せる。観測系はint64。writeの点ではinterfaceに寄せてint
	return n, err
}

// flush クライアントに即データ返す
// 今は別にflush使わんけどstreamingとかで使う。その時にpanic起こさないように保険
/*responseRecorderはflush実装してるからこれがhandlerに渡れば当然flushは入れ子的に満たされるが、テストで差し替えたり
別middleware通って実態がresponseRecorderじゃなくなった場合それがflush持ってるとは限らない。だからf, ok := r.w.(http.Flusher)
型アサーションで確認してから実行*/
//responseRecorderは当然中身が代入される->f, ok := r.w.(http.Flusher)の意味ある->腹落ち

func (r *responseRecorder) Flush() {
	if f, ok := r.w.(http.Flusher); ok {
		f.Flush()
	}
}

//Hijack(webscoket)/Push等が必要になったらここに透過実装

func AccessLog(lg logger.Logger) func(http.Handler) http.Handler { //これ典型記法
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rec := &responseRecorder{
				w:          w,
				statusCode: http.StatusOK, //デフォルトを定義
			}

			next.ServeHTTP(rec, r)
			//handler/router/usecaseがここで全部実行される。
			//ちゃうわ。普通にここが起点というより起点の伝搬。前後でかかる時間測定
			//ここでさっき定義したwriteheader使う。

			dur := time.Since(start)

			attrs := []logger.Attr{
				//request_idはloggerの具体でつけてる
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path), // queryはログしない。tokenやsession系が混ざったりそもそも検索しづらい。
				logger.Int("status", rec.statusCode),
				logger.Int64("bytes", rec.bytes),                //観測系はint64
				logger.Int64("duration_us", dur.Microseconds()), //millisecondやとちょっと粗い。0msが多発する。
				logger.String("ip", clientIP(r)),
				logger.String("ua", userAgent(r)),
			} //比較、計算したい奴はintで返す。metrics、アラート条件で使う。

			if rec.statusCode >= 500 {
				lg.Error(r.Context(), "access", nil, attrs...) //accesslogもlogに入れてaccessキーワードで見る。
				//エラー本体は返さない、
				return
				//ErrorはアラートなるからaccesslogはInfo固定にする場合もある
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

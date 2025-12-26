package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

// 取得をここに定義したくないからwithvalueの埋め込みと取得のヘルパーをobservability/requestidに移行
const HeaderRequestID = "X-Request-Id"

/* FromContext は logger 等から request_id を取得するための helper
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}ここで定義するとloggerが終わる*/

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 既存のrequest idを尊重
		rid := r.Header.Get(HeaderRequestID)
		//TODO:長すぎたら捨てて生成し直す
		if rid == "" {
			rid = uuid.NewString()
		}

		ctx := requestid.WithContext(r.Context(), rid)

		w.Header().Set(HeaderRequestID, rid)
		//request_idはユーザーも使う。request_idでお問い合わせ。
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/*package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKeyRequestID struct{}

var requestIDKey = ctxKeyRequestID{}

// これは衝突を避けるため。string、type ctxKeyRequestID stringは不可
const HeaderRequestID = "X-Request-Id"

/* FromContext は logger 等から request_id を取得するための helper
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}ここで定義するとloggerが終わる*/

//internal/observability/requestid
//ここはerrで返してない。つまりあるか確認するだけ。
// middlewareで保証してるけどここでもerrで保険は良くない。
// べつにidがなくてもロジックには問題ないゆえに保険かけない。
//仮にrequestidがロジックに入ってるならdomainVOとか別の境界で保証する保険あってもいい。

/*func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 既存の request id を尊重
		rid := r.Header.Get(HeaderRequestID)
		//TODO:長すぎたら捨てて生成し直す
		if rid == "" {
			rid = uuid.NewString()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, rid)

		w.Header().Set(HeaderRequestID, rid)
		//request_idはユーザーも使う。request_idでお問い合わせ。
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
*/

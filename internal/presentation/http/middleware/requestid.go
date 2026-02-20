package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

const HeaderRequestID = "X-Request-Id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rid := r.Header.Get(HeaderRequestID)
		//TODO:長すぎたら捨てて生成し直す

		if rid == "" {
			rid = uuid.NewString()
		}

		ctx := requestid.WithContext(r.Context(), rid)

		w.Header().Set(HeaderRequestID, rid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

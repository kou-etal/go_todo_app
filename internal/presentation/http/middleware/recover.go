package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
)

func Recover(lg logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if p := recover(); p != nil {
					lg.Error(r.Context(), "panic recovered",
						fmt.Errorf("%v", p),
						logger.String("stack", string(debug.Stack())),
					)
					responder.JSON(w, http.StatusInternalServerError,
						responder.ErrResponse{Message: "internal server error"},
					)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

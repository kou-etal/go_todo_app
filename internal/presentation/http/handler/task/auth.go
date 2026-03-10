package task

import (
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/auth"
	"github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
)

func userIDFromRequest(w http.ResponseWriter, r *http.Request) (user.UserID, bool) {
	uid, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "unauthorized"})
		return "", false
	}
	return uid, true
}

package auth

import (
	"context"

	"github.com/kou-etal/go_todo_app/internal/domain/user"
)

type ctxKeyUserID struct{}

var userIDKey = ctxKeyUserID{}

func WithUserID(ctx context.Context, uid user.UserID) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func UserIDFromContext(ctx context.Context) (user.UserID, bool) {
	id, ok := ctx.Value(userIDKey).(user.UserID)
	return id, ok
}

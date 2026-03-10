package middleware

import (
	"net/http"
	"strings"

	"github.com/kou-etal/go_todo_app/internal/auth"
	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
)

func Auth(parser auth.AccessTokenParser, lg logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			header := r.Header.Get("Authorization")
			if header == "" {
				responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "authorization header is required"})
				return
			}

			token, found := strings.CutPrefix(header, "Bearer ")
			if !found || token == "" {
				responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "invalid authorization header format"})
				return
			}

			userID, err := parser.ParseAccessToken(token)
			if err != nil {
				lg.Debug(ctx, "invalid access token", err)
				responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "invalid or expired token"})
				return
			}

			uid, err := duser.ParseUserID(userID)
			if err != nil {
				lg.Error(ctx, "invalid user id in token", err)
				responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "invalid token"})
				return
			}

			ctx = auth.WithUserID(ctx, uid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

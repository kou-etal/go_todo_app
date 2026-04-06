package userhandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/refresh"
)

type refreshHandler struct {
	uc     *refresh.Usecase
	logger logger.Logger
}

func NewRefresh(uc *refresh.Usecase, lg logger.Logger) *refreshHandler {
	return &refreshHandler{uc: uc, logger: lg}
}

func (h *refreshHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req refreshRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		h.logger.Debug(ctx, "invalid json body", err)
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid body"})
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.logger.Debug(ctx, "invalid json body: trailing data", nil)
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid body"})
		return
	}

	cmd := refresh.Command{
		RefreshToken: req.RefreshToken,
	}

	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		case errors.Is(err, refresh.ErrEmptyRefreshToken):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "refresh token is required"})
			return
		case errors.Is(err, refresh.ErrInvalidRefreshToken):
			responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "invalid or expired refresh token"})
			return
		default:
			h.logger.Error(ctx, "refresh failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}

	responder.JSON(w, http.StatusOK, refreshResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

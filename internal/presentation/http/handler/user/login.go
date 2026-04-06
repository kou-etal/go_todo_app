package userhandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/login"
)

type loginHandler struct {
	uc     *login.Usecase
	logger logger.Logger
}

func NewLogin(uc *login.Usecase, lg logger.Logger) *loginHandler {
	return &loginHandler{uc: uc, logger: lg}
}

func (h *loginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req loginRequest
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

	cmd := login.Command{
		Email:    req.Email,
		Password: req.Password,
	}

	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		case errors.Is(err, login.ErrEmptyEmail):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "email is required"})
			return
		case errors.Is(err, login.ErrEmptyPassword):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "password is required"})
			return
		case errors.Is(err, login.ErrInvalidCredentials):
			responder.JSON(w, http.StatusUnauthorized, responder.ErrResponse{Message: "invalid credentials"})
			return
		default:
			h.logger.Error(ctx, "login failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}

	responder.JSON(w, http.StatusOK, loginResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresIn:    res.ExpiresIn,
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

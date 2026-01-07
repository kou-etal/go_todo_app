package userhandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	duser "github.com/kou-etal/go_todo_app/internal/domain/user"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/register"
)

type RegisterUserHandler struct {
	uc     *register.Usecase
	logger logger.Logger
}

func NewRegister(uc *register.Usecase, lg logger.Logger) *RegisterUserHandler {
	return &RegisterUserHandler{
		uc:     uc,
		logger: lg,
	}
}

func (h *RegisterUserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req registerRequest
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

	cmd := register.Command{
		Email:    req.Email,
		Password: req.Password,
		UserName: req.UserName,
	}

	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		//TODO:エラー設計雑すぎ、そもそもhandlerがdomainのエラー扱わない。usecaseで吸収する。
		// usecaseエラー
		case errors.Is(err, register.ErrEmptyEmail),
			errors.Is(err, register.ErrEmptyPassword),
			errors.Is(err, register.ErrEmptyUserName):
			h.logger.Debug(ctx, "invalid command", nil)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid request"})
			return

		//domainエラー
		case errors.Is(err, duser.ErrEmptyEmail),
			errors.Is(err, duser.ErrEmailTooLong),
			errors.Is(err, duser.ErrInvalidEmailFormat),
			errors.Is(err, duser.ErrEmptyPassword),
			errors.Is(err, duser.ErrPasswordTooShort),
			errors.Is(err, duser.ErrPasswordTooLong),
			errors.Is(err, duser.ErrPasswordHasLeadingOrTrailingSpace),
			errors.Is(err, duser.ErrEmptyName),
			errors.Is(err, duser.ErrNameTooLong):
			h.logger.Debug(ctx, "domain validation failed", err)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid request"})
			return

		default:
			h.logger.Error(ctx, "register user failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}

	responder.JSON(w, http.StatusCreated, registerResponse{
		ID: res.UserID,
	})
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	UserName string `json:"user_name"`
} //TODO:あれjsonってsnake_caseかcamelCaseどっちやっけ

type registerResponse struct {
	ID string `json:"id"`
}

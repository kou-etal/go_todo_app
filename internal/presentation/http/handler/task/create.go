package task

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
)

type createHandler struct {
	uc     *create.Usecase
	logger logger.Logger
}

func NewCreate(uc *create.Usecase, lg logger.Logger) *createHandler {
	return &createHandler{
		uc:     uc,
		logger: lg,
	}
}
func (h *createHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		h.logger.Debug(
			ctx,
			"invalid json body",
			err,
		)
		responder.JSON(
			w,
			http.StatusBadRequest,
			responder.ErrResponse{Message: "invalid body"},
		)
		return
	}

	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.logger.Debug(
			ctx,
			"invalid json body: trailing data",
			nil,
		)
		responder.JSON(
			w,
			http.StatusBadRequest,
			responder.ErrResponse{Message: "invalid body"},
		)
		return
	}

	cmd := create.Command{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
	}
	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		case errors.Is(err, create.ErrEmptyTitle):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "title"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "title is required"})
			return
		case errors.Is(err, create.ErrTitleTooLong):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "title"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "title is too long"})
			return
		case errors.Is(err, create.ErrEmptyDescription):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "description"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "description is required"})
			return
		case errors.Is(err, create.ErrDescriptionTooLong):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "description"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "description is too long"})
			return
		case errors.Is(err, create.ErrInvalidDueOption):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "due_date"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid due_date"})
			return
		default:
			h.logger.Error(ctx, "create task failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}
	responder.JSON(w, http.StatusCreated, createResponse{ID: res.ID})
}

type createRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	DueDate     int    `json:"due_date"`
}

type createResponse struct {
	ID string `json:"id"`
}

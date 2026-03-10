package task

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
)

type deleteHandler struct {
	uc     *remove.Usecase
	logger logger.Logger
}

func NewDelete(uc *remove.Usecase, lg logger.Logger) *deleteHandler {
	return &deleteHandler{
		uc:     uc,
		logger: lg,
	}
}
func (h *deleteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uid, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	if id == "" {

		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
			Message: "invalid id",
		})
		return
	}
	var req deleteRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		h.logger.Debug(ctx, "invalid json body", err)
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
			Message: "invalid body",
		})
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		h.logger.Debug(ctx, "invalid json body: trailing data", nil)
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
			Message: "invalid body",
		})
		return
	}

	cmd := remove.Command{
		UserID:  uid.Value(),
		ID:      id,
		Version: req.Version,
	}
	if err := h.uc.Do(ctx, cmd); err != nil {
		switch {
		case errors.Is(err, remove.ErrInvalidID):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid id"})
			return
		case errors.Is(err, remove.ErrInvalidVersion):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid version"})
			return
		case errors.Is(err, remove.ErrNotFound):
			responder.JSON(w, http.StatusNotFound, responder.ErrResponse{Message: "not found"})
			return
		case errors.Is(err, remove.ErrConflict):
			responder.JSON(w, http.StatusConflict, responder.ErrResponse{Message: "conflict"})
			return
		default:
			h.logger.Error(ctx, "delete task failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

type deleteRequest struct {
	Version uint64 `json:"version"`
}

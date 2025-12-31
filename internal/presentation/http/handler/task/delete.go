package taskhandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
)

// versionをheaderかbodyどっちに含ませるか議論
// If-Match(headerに含ませる)方がhttp的に正しい
// 今回はbodyに含ませる設計
type DeleteTaskHandler struct {
	uc     *delete.Usecase
	logger logger.Logger
}

func NewDelete(uc *delete.Usecase, lg logger.Logger) *DeleteTaskHandler {
	return &DeleteTaskHandler{
		uc:     uc,
		logger: lg,
	}
}
func (h *DeleteTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		//ここはDebugなし。別にinternalエラーでもない。再現可能
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

	cmd := delete.Command{
		ID:      id,
		Version: req.Version,
	} //usecaseエラー
	if err := h.uc.Do(ctx, cmd); err != nil {
		switch {
		case errors.Is(err, delete.ErrInvalidID),
			errors.Is(err, delete.ErrInvalidVersion):
			h.logger.Debug(ctx, "invalid command", nil)
			//TODO:ここはdebugエラー返さずにfieldとreason返す
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
				Message: "invalid request",
			})
			return

			//repoで発生したエラー、domainに寄せてる。
		case errors.Is(err, dtask.ErrConflict):
			responder.JSON(w, http.StatusConflict, responder.ErrResponse{
				Message: "conflict",
			})
			return

		default:
			h.logger.Error(ctx, "delete task failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{
				Message: "internal server error",
			})
			return
		}
	}
	w.WriteHeader(http.StatusNoContent) //bodyなしゆえにresponder.JSON使わない。
}

type deleteRequest struct {
	Version uint64 `json:"version"`
}

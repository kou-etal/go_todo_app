package task

//package taskhandlerこれはhandler二重

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
)

// versionをheaderかbodyどっちに含ませるか議論
// If-Match(headerに含ませる)方がhttp的に正しい
// 今回はbodyに含ませる設計
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

	cmd := remove.Command{
		ID:      id,
		Version: req.Version,
	} //usecaseエラー
	if err := h.uc.Do(ctx, cmd); err != nil {
		switch {
		//TODO:エラー設計雑すぎ、そもそもhandlerがdomainのエラー扱わない。usecaseで吸収する。
		case errors.Is(err, remove.ErrInvalidID),
			errors.Is(err, remove.ErrInvalidVersion):
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

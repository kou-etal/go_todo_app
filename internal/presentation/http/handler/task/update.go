package taskhandler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
)

type UpdateTaskHandler struct {
	uc     *update.Usecase
	logger logger.Logger
}

func NewUpdate(uc *update.Usecase, lg logger.Logger) *UpdateTaskHandler {
	return &UpdateTaskHandler{
		uc:     uc,
		logger: lg,
	}
}

func (h *UpdateTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := r.PathValue("id")
	if id == "" {
		//ここはDebugなし。別にinternalエラーでもない。再現可能
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
			Message: "invalid id",
		})
		return
	}

	var req updateRequest
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) //DOS対策。1MB。wに記述。

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&req); err != nil {
		h.logger.Debug(
			ctx,
			"invalid json body",
			err, //ここはdebugでエラー返す
		)
		responder.JSON(
			w,
			http.StatusBadRequest,
			responder.ErrResponse{Message: "invalid body"},
		)
		return
	}
	//{ "title": "a" }{ "title": "b" }を防ぐ
	//&struct{}{}はtype struct a{}  dec.Decode(&a{});
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

	cmd := update.Command{
		ID:          id,
		Version:     req.Version,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate,
	} //なしはnil
	//TODO:listでversion返す

	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		case errors.Is(err, update.ErrInvalidID),
			errors.Is(err, update.ErrInvalidVersion),
			errors.Is(err, update.ErrNoFieldsToUpdate),
			errors.Is(err, update.ErrInvalidTitle),
			errors.Is(err, update.ErrInvalidDescription),
			errors.Is(err, update.ErrInvalidDueOption):
			h.logger.Debug(ctx, "invalid command", nil)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
				Message: "invalid request",
			})
			return //usecaseエラー
			//TODO:ここはdebugエラー返さずにfieldとreason返す

		case errors.Is(err, dtask.ErrEmptyTitle),
			errors.Is(err, dtask.ErrTitleTooLong),
			errors.Is(err, dtask.ErrEmptyDescription),
			errors.Is(err, dtask.ErrDescriptionTooLong),
			errors.Is(err, dtask.ErrInvalidDueOption):
			h.logger.Debug(ctx, "domain validation failed", nil)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
				Message: "invalid request",
			})
			return
			//TODO:ここはdebugエラー返さずにfieldとreason返す
		case errors.Is(err, dtask.ErrNotFound):
			responder.JSON(w, http.StatusNotFound, responder.ErrResponse{
				Message: "not found",
			})
			return
			//ここはdebug返さない
		case errors.Is(err, dtask.ErrConflict):
			responder.JSON(w, http.StatusConflict, responder.ErrResponse{
				Message: "conflict",
			})
			return
			//ここはdebug返さない
		default:
			h.logger.Error(ctx, "update task failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{
				Message: "internal server error",
			})
			return
		} //ここはinternalエラー

	}
	responder.JSON(w, http.StatusOK, updateResponse{
		ID: res.ID,
	})
}

type updateRequest struct {
	Version     uint64  `json:"version"`
	Title       *string `json:"title,omitempty"` //0とnilを分離
	Description *string `json:"description,omitempty"`
	DueDate     *int    `json:"due_date,omitempty"` //jsonはjs onlyでもない限りsnake_case
}

// 今の更新の設計はpatch系。変わらない分は更新しない。put系と比べて意味ない更新が起こらないメリット。
type updateResponse struct {
	ID string `json:"id"`
	//TODO:version返す
}

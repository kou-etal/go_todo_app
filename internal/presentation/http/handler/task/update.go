package task

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
)

type updateHandler struct {
	uc     *update.Usecase
	logger logger.Logger
}

func NewUpdate(uc *update.Usecase, lg logger.Logger) *updateHandler {
	return &updateHandler{
		uc:     uc,
		logger: lg,
	}
}

func (h *updateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if req.Version == 0 { //これするんやったらversionを1始まり。
		//TODO:いやでもポインタにしてnilで判定する方が安全。0は別に送れてしまう。
		responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
			Message: "invalid version",
		})
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
	//commandへ完全raw data送るわけではない。id,versionないとかおかしすぎるやつはhandlerで弾く

	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		case errors.Is(err, update.ErrInvalidID):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid id"})
			return
		case errors.Is(err, update.ErrInvalidVersion):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid version"})
			return
		case errors.Is(err, update.ErrNoFieldsToUpdate):
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "no fields to update"})
			return
		case errors.Is(err, update.ErrEmptyTitle):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "title"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "title is required"})
			return
		case errors.Is(err, update.ErrTitleTooLong):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "title"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "title is too long"})
			return
		case errors.Is(err, update.ErrEmptyDescription):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "description"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "description is required"})
			return
		case errors.Is(err, update.ErrDescriptionTooLong):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "description"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "description is too long"})
			return
		case errors.Is(err, update.ErrInvalidDueOption):
			h.logger.Debug(ctx, "invalid command", nil, logger.String("field", "due_date"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid due_date"})
			return
		case errors.Is(err, update.ErrNotFound):
			responder.JSON(w, http.StatusNotFound, responder.ErrResponse{Message: "not found"})
			return
		case errors.Is(err, update.ErrConflict):
			responder.JSON(w, http.StatusConflict, responder.ErrResponse{Message: "conflict"})
			return
		default:
			h.logger.Error(ctx, "update task failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
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

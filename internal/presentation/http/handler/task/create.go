package task

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
)

type CreateTaskHandler struct {
	uc     *create.Usecase
	logger logger.Logger
}

func NewCreate(uc *create.Usecase, lg logger.Logger) *CreateTaskHandler {
	return &CreateTaskHandler{
		uc:     uc,
		logger: lg,
	}
}
func (h *CreateTaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req createRequest
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

	cmd := create.Command{
		Title:       req.Title,
		Description: req.Description,
		DueDate:     req.DueDate, //7/14/21/30
	} //いやでもdtask.dueoption変換の責務をcomandが持たなかったらcommandの意味なくねっていう考え方もある。
	res, err := h.uc.Do(ctx, cmd)
	if err != nil {
		switch {
		//TODO:エラーメッセージが雑すぎる。Message: "invalid request"は不親切。
		// TODO:あとdomainのエラーはusecaseで吸収したほうがいい。今やとdomain変更したらhandlerも変更。
		// usecaseエラー
		//ここはdebugでエラー返さなくていい。fieldとreason返す。
		case errors.Is(err, create.ErrInvalidTitle):
			h.logger.Debug(
				ctx,
				"invalid command",
				nil,
				logger.String("field", "title"),
				logger.String("reason", "invalid"),
			)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
				Message: "invalid request",
			})
			return
		case errors.Is(err, create.ErrInvalidDescription):
			h.logger.Debug(
				ctx,
				"invalid command",
				nil,
				logger.String("field", "description"),
				logger.String("reason", "invalid"),
			)
			responder.JSON(
				w,
				http.StatusBadRequest,
				responder.ErrResponse{Message: "invalid request"},
			)
			return
		case errors.Is(err, create.ErrInvalidDueOption):
			h.logger.Debug(
				ctx,
				"invalid command",
				nil,
				logger.String("field", "due_date"),
				logger.String("reason", "invalid"),
			)
			responder.JSON(
				w,
				http.StatusBadRequest,
				responder.ErrResponse{Message: "invalid request"},
			)
			return
		//400系はDebugで返す。
		case errors.Is(err, dtask.ErrEmptyTitle),
			errors.Is(err, dtask.ErrTitleTooLong),
			errors.Is(err, dtask.ErrEmptyDescription),
			errors.Is(err, dtask.ErrDescriptionTooLong),
			errors.Is(err, dtask.ErrInvalidDueOption):
			h.logger.Debug(ctx, "domain validation failed", err)
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid request"})
			return
		//domainエラー
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
	DueDate     int    `json:"due_date"` // 7/14/21/30
}

type createResponse struct {
	ID string `json:"id"`
}

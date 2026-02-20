package task

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
)

type listHandler struct {
	uc     *list.Usecase
	logger logger.Logger
}

func NewList(uc *list.Usecase, lg logger.Logger) *listHandler {
	return &listHandler{
		uc:     uc,
		logger: lg,
	}
}

func (h *listHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qp := r.URL.Query()

	var q list.Query
	q.Sort = qp.Get("sort")
	q.Cursor = qp.Get("cursor")

	if v := qp.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {

			h.logger.Debug(ctx, "invalid limit", nil, logger.String("limit", v))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{
				Message: "invalid limit",
			})

			return
		}
		q.Limit = limit
	}

	res, err := h.uc.Do(ctx, q)
	if err != nil {
		switch {
		case errors.Is(err, list.ErrInvalidLimit):
			h.logger.Debug(ctx, "invalid query", nil, logger.String("field", "limit"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid limit"})
			return
		case errors.Is(err, list.ErrInvalidSort):
			h.logger.Debug(ctx, "invalid query", nil, logger.String("field", "sort"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid sort"})
			return
		case errors.Is(err, list.ErrInvalidCursor):
			h.logger.Debug(ctx, "invalid query", nil, logger.String("field", "cursor"))
			responder.JSON(w, http.StatusBadRequest, responder.ErrResponse{Message: "invalid cursor"})
			return
		default:
			h.logger.Error(ctx, "list tasks failed", err)
			responder.JSON(w, http.StatusInternalServerError, responder.ErrResponse{Message: "internal server error"})
			return
		}
	}

	out := toHTTPListResponse(res)

	responder.JSON(w, http.StatusOK, out)
}

type listResponse struct {
	Items      []listItem `json:"items"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

type listItem struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description,omitempty"`
	Status      string  `json:"status"`
	DueDate     *string `json:"due_date,omitempty"`
}

func toHTTPListResponse(res list.Result) listResponse {
	items := make([]listItem, 0, len(res.Items))
	for _, it := range res.Items {
		var due *string
		if it.DueDate != nil {
			s := it.DueDate.UTC().Format(timeRFC3339)
			due = &s
		}
		items = append(items, listItem{
			ID:          it.ID,
			Title:       it.Title,
			Description: it.Description,
			Status:      it.Status,
			DueDate:     due,
		})
	}

	return listResponse{
		Items:      items,
		NextCursor: res.NextCursor,
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

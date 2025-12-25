package taskhandler

//これは中規模ならpackage handlerで統一してもいい

import (
	"errors"
	"net/http"
	"strconv"

	dtask "github.com/kou-etal/go_todo_app/internal/domain/task"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/responder"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
)

/*handlerのテストは薄いから実務ではinterface経由しないこと多い
type ListTasksUsecase interface {
	Execute(ctx context.Context, q list.Query) (list.Result, error)
}
type ListTasksHandler struct {
	uc ListTasksUsecase
}*/
// internal/presentation/http/handler/task/list.go
type ListTasksHandler struct {
	uc     *list.Usecase
	logger logger.Logger
}

// NewListTasksHandlerではない
func New(uc *list.Usecase, lg logger.Logger) *ListTasksHandler {
	return &ListTasksHandler{
		uc:     uc,
		logger: lg,
	}
}

func (h *ListTasksHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	qp := r.URL.Query()

	var q list.Query
	q.Sort = qp.Get("sort")
	q.Cursor = qp.Get("cursor")

	if v := qp.Get("limit"); v != "" {
		limit, err := strconv.Atoi(v)
		if err != nil {
			//4xxはログ欲しければDebugに落とす。機密注意。
			h.logger.Debug(ctx, "invalid limit", logger.String("limit", v))
			responder.JSON(w, http.StatusBadRequest, errResponse{
				Message: "invalid limit",
			})
			//入力不正
			return
		}
		q.Limit = limit
	}

	res, err := h.uc.Do(ctx, q)
	if err != nil {
		switch {
		case errors.Is(err, dtask.ErrInvalidCursor),
			errors.Is(err, dtask.ErrInvalidLimit),
			errors.Is(err, dtask.ErrInvalidSort):
			h.logger.Debug(ctx, "invalid query")
			//Attrは可変長引数　0でもok
			responder.JSON(w, http.StatusBadRequest, errResponse{
				Message: "invalid query",
			})
			//TODO:セキュリティ配慮とクライアント体験のトレードオフ、どこまでエラー返すか考える
			//入力不正
			//入力不正系はwrapしないで返すのに。であたかもwrapしてるかのように吸収する。保険

			return
		default:
			//プログラムの破綻
			h.logger.Error(ctx, "list tasks failed", err)
			responder.JSON(w, http.StatusInternalServerError, errResponse{
				Message: "internal server error",
			})
			return
		}
	}

	out := toHTTPListResponse(res)

	responder.JSON(w, http.StatusOK, out)
}

type errResponse struct {
	Message string `json:"message"`
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

// entity->result->http
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

//responderは一つで統一(writeerr作らない)

package router

import (
	"net/http"
)

/*type Deps struct {
	TaskList *taskhandler.ListTasksHandler
}*/

type Deps struct {
	Task TaskDeps
}

type TaskDeps struct {
	List   http.Handler
	Create http.Handler
	Update http.Handler
	Delete http.Handler
}

// http.Handlerはinterface
// これによりtaskhandler "github.com/kou-etal/go_todo_app/internal/presentation/http/handler/task"が消え疎結合になった
func New(deps Deps) http.Handler {
	mux := http.NewServeMux()

	// health
	mux.Handle("/health", http.HandlerFunc(healthHandler))

	// tasks
	mux.Handle("/tasks", deps.Task.List)
	//h := middleware.RequestID(mux)  middlewareのチェーンはここでやると汚い

	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

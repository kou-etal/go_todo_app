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

	//tasksHandlerはList返すだけではなくservehttp実行まで行う。実行がtasksHandlerに使われるから。
	//mux.Handle("/tasks", deps.Task.List))は可能やけどそれは実行がdeps.Task.Listに使われるから。

	//RESTに従うためにパスではなくメソッドで分岐
	/*両方を分岐させるtasksHandlerを作った場合GET /task/123とかがはじけない。それを弾こうと思うとurlをrouterで取得することになるが
	今の設計ではurl取得をhandlerに寄せてるから不適*/
	//collection
	mux.Handle("/tasks", tasksHandler(deps.Task))
	//item
	mux.Handle("/tasks/", taskHandler(deps.Task))
	//h := middleware.RequestID(mux)  middlewareのチェーンはここでやると汚い

	return mux
}

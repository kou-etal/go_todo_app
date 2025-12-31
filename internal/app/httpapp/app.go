// DI層　app->router->handler->usecase->domain->repoの依存

package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/clock"
	"github.com/kou-etal/go_todo_app/internal/config"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	taskrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/task"

	"github.com/kou-etal/go_todo_app/internal/logger"

	taskhandler "github.com/kou-etal/go_todo_app/internal/presentation/http/handler/task"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/middleware"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/router"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
)

// Buildは依存を組み立ててhttp.Handlerとcleanupを返す。
// mainはserver起動とshutdownだけやればいい。
// errはmainで受け取ってlogger。build側はあくまでwiring
func Build(ctx context.Context) (http.Handler, func(), error) {
	cfg, err := config.New()
	if err != nil {
		return nil, func() {}, fmt.Errorf("get config: %w", err)
	}
	clk := clock.RealClocker{}
	lg := logger.NewSlog()
	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return nil, func() {}, fmt.Errorf("start mysql: %w", err)
	}

	cleanup := func() {
		closeDB()
	}

	taskRepo := taskrepo.NewRepository(xdb)
	//taskrepo.NewRepositoryはqueryer、これはもともとsqlxが満たしているメソッド。
	//重い抽象ではない軽い抽象

	tasklistUC := list.New(taskRepo)
	taskCreateUC := create.New(taskRepo, clk)
	taskUpdateUC := update.New(taskRepo, clk)
	taskDeleteUC := remove.New(taskRepo)

	taskListHandler := taskhandler.NewList(tasklistUC, lg)
	taskCreateHandler := taskhandler.NewCreate(taskCreateUC, lg)
	taskUpdateHandler := taskhandler.NewUpdate(taskUpdateUC, lg)
	taskDeleteHandler := taskhandler.NewDelete(taskDeleteUC, lg)

	h := router.New(router.Deps{
		Task: router.TaskDeps{
			List:   taskListHandler,
			Create: taskCreateHandler,
			Update: taskUpdateHandler,
			Delete: taskDeleteHandler,
		},
	})

	h = middleware.RequestID(h)
	// h = middleware.Recover(lg)(h)
	h = middleware.AccessLog(lg)(h)

	return h, cleanup, nil
}

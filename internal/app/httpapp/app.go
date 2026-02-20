package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/kou-etal/go_todo_app/internal/app/txdeps"
	"github.com/kou-etal/go_todo_app/internal/clock"
	"github.com/kou-etal/go_todo_app/internal/config"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	txrunner "github.com/kou-etal/go_todo_app/internal/infra/db/tx"
	emailverifyrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/emailverify"
	taskeventrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/event"
	taskrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/task"
	userrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/user"
	"github.com/kou-etal/go_todo_app/internal/infra/security"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/handler/task"
	userhandler "github.com/kou-etal/go_todo_app/internal/presentation/http/handler/user"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/middleware"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/router"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/register"
)

func Build(ctx context.Context, cfg *config.Config) (http.Handler, func(), error) {

	clk := clock.RealClocker{}
	lg := logger.NewSlog()
	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return nil, func() {}, fmt.Errorf("start mysql: %w", err)
	}

	cleanup := func() {
		closeDB()
	}

	taskRepo := taskrepo.New(xdb)

	taskListUC := list.New(taskRepo)
	taskListHandler := task.NewList(taskListUC, lg)

	makeRegisterDeps := func(q db.QueryerExecer) usetx.RegisterDeps {
		u := userrepo.NewRepository(q)
		v := emailverifyrepo.NewRepository(q)
		return txdeps.NewRegister(u, v)
	}

	beginner := xdb

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}

	registerRunner := txrunner.New(beginner, txOpts, makeRegisterDeps)
	passwordHasher := security.NewBcryptHasher(14)
	tokenGenerator := security.NewRandomTokenGenerator(32)
	tokenHasher := security.SHA256TokenHasher{}
	registerUC := register.New(
		registerRunner,
		clk,
		passwordHasher,
		tokenGenerator,
		tokenHasher,
	)

	makeTaskEventDeps := func(q db.QueryerExecer) usetx.TaskEventDeps {
		t := taskrepo.New(q)
		e := taskeventrepo.New(q)
		return txdeps.NewTaskEvent(t, e)
	}
	taskEventRunner := txrunner.New[usetx.TaskEventDeps](beginner, txOpts, makeTaskEventDeps)
	taskCreateUC := create.New(taskEventRunner, clk)
	taskCreateHandler := task.NewCreate(taskCreateUC, lg)

	taskUpdateUC := update.New(taskEventRunner, clk)
	taskDeleteUC := remove.New(taskEventRunner, clk)
	taskUpdateHandler := task.NewUpdate(taskUpdateUC, lg)
	taskDeleteHandler := task.NewDelete(taskDeleteUC, lg)

	userRegisterHandler := userhandler.NewRegister(registerUC, lg)

	h := router.New(router.Deps{
		Task: router.TaskDeps{
			List:   taskListHandler,
			Create: taskCreateHandler,
			Update: taskUpdateHandler,
			Delete: taskDeleteHandler,
		},
		User: router.UserDeps{
			Register: userRegisterHandler,
		},
	})
	//TODO:middlewareチェーンはrouter/middlewareに託す
	h = middleware.RequestID(h)
	// h = middleware.Recover(lg)(h)
	h = middleware.AccessLog(lg)(h)

	return h, cleanup, nil
}

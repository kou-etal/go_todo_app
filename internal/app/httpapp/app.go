package app

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/kou-etal/go_todo_app/internal/app/txdeps"
	"github.com/kou-etal/go_todo_app/internal/auth"
	"github.com/kou-etal/go_todo_app/internal/clock"
	"github.com/kou-etal/go_todo_app/internal/config"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	txrunner "github.com/kou-etal/go_todo_app/internal/infra/db/tx"
	emailverifyrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/emailverify"
	taskeventrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/event"
	refreshrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/refresh"
	taskrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/task"
	userrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/user"
	"github.com/kou-etal/go_todo_app/internal/infra/security"

	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
	oteltrace "github.com/kou-etal/go_todo_app/internal/observability/trace"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/handler/task"
	userhandler "github.com/kou-etal/go_todo_app/internal/presentation/http/handler/user"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/middleware"
	"github.com/kou-etal/go_todo_app/internal/presentation/http/router"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/create"
	remove "github.com/kou-etal/go_todo_app/internal/usecase/task/delete"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/list"
	"github.com/kou-etal/go_todo_app/internal/usecase/task/update"
	usetx "github.com/kou-etal/go_todo_app/internal/usecase/tx"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/login"
	userefresh "github.com/kou-etal/go_todo_app/internal/usecase/user/refresh"
	"github.com/kou-etal/go_todo_app/internal/usecase/user/register"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Build(ctx context.Context, cfg *config.Config) (http.Handler, http.Handler, func(), error) {

	clk := clock.RealClocker{}
	lg := logger.NewSlog()

	var shutdownTracer func(context.Context) error
	//ごみ掃除関数。
	if cfg.OTLPEndpoint != "" {
		tp, err := oteltrace.NewProvider(ctx, cfg.ServiceName, cfg.OTLPEndpoint)
		//provider宣言。
		if err != nil {
			return nil, nil, func() {}, fmt.Errorf("init tracer: %w", err)
		}
		shutdownTracer = tp.Shutdown //ifの外でshutdownを使うために置く。
	}

	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return nil, nil, func() {}, fmt.Errorf("start mysql: %w", err)
	}

	cleanup := func() {
		closeDB()
		if shutdownTracer != nil {
			if err := shutdownTracer(context.Background()); err != nil {
				slog.Error("tracer shutdown failed", slog.String("error", err.Error()))
			}
		}
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
	passwordHasher := security.NewBcryptHasher(12)
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

	// JWT
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.AccessTokenTTL)
	refreshTTL := time.Duration(cfg.RefreshTokenTTL) * time.Second

	// Login
	makeLoginDeps := func(q db.QueryerExecer) usetx.LoginDeps {
		u := userrepo.NewRepository(q)
		r := refreshrepo.NewRepository(q)
		return txdeps.NewLogin(u, r)
	}
	loginRunner := txrunner.New(beginner, txOpts, makeLoginDeps)
	loginUC := login.New(
		loginRunner,
		clk,
		passwordHasher,
		jwtManager,
		tokenHasher,
		tokenGenerator,
		refreshTTL,
		cfg.AccessTokenTTL,
	)

	// Refresh
	makeRefreshDeps := func(q db.QueryerExecer) usetx.RefreshDeps {
		r := refreshrepo.NewRepository(q)
		return txdeps.NewRefresh(r)
	}
	refreshRunner := txrunner.New(beginner, txOpts, makeRefreshDeps)
	refreshUC := userefresh.New(
		refreshRunner,
		clk,
		jwtManager,
		tokenHasher,
		tokenGenerator,
		refreshTTL,
		cfg.AccessTokenTTL,
	)

	userRegisterHandler := userhandler.NewRegister(registerUC, lg)
	userLoginHandler := userhandler.NewLogin(loginUC, lg)
	userRefreshHandler := userhandler.NewRefresh(refreshUC, lg)

	authMW := middleware.Auth(jwtManager, lg)

	h := router.New(router.Deps{
		Task: router.TaskDeps{
			List:   taskListHandler,
			Create: taskCreateHandler,
			Update: taskUpdateHandler,
			Delete: taskDeleteHandler,
		},
		User: router.UserDeps{
			Register: userRegisterHandler,
			Login:    userLoginHandler,
			Refresh:  userRefreshHandler,
		},
		AuthMW: authMW,
	})
	h = middleware.RequestID(h)
	h = middleware.Recover(lg)(h)
	h = middleware.AccessLog(lg)(h)

	// Metrics
	mp := metrics.NewProvider()
	mp.Registry.MustRegister(collectors.NewDBStatsCollector(xdb.DB, "todo_db"))
	httpMetrics := metrics.NewHTTPMetrics(mp.Registry)
	h = middleware.Metrics(httpMetrics)(h)

	if cfg.OTLPEndpoint != "" { //エンドポイント存在する場合はwrapする。
		h = otelhttp.NewHandler(h, "todo-api")
	}

	if cfg.CORSOrigin != "" {
		h = middleware.CORS(cfg.CORSOrigin)(h)
	}

	return h, mp.Handler(), cleanup, nil
}

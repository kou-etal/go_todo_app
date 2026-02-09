// DI層　app->router->handler->usecase->domain->repoの依存
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

// Buildは依存を組み立ててhttp.Handlerとcleanupを返す。
// mainはserver起動とshutdownだけやればいい。
// errはmainで受け取ってlogger。build側はあくまでwiring
func Build(ctx context.Context, cfg *config.Config) (http.Handler, func(), error) {
	//mainとapp両方config使うからmainでnew作ってappに与える
	clk := clock.RealClocker{}
	lg := logger.NewSlog()
	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return nil, func() {}, fmt.Errorf("start mysql: %w", err)
	}

	cleanup := func() {
		closeDB()
	}

	/*ここのの命名は怪しい
	まず繰り返しをしない、長い名前を避けるは大事
	taskとrepoはpackage taskにしてrepoをエイリアス、usecaseは行動単位をパッケージにする
	handlerを省略すべきか
	*/
	taskRepo := taskrepo.New(xdb)

	//taskrepo.NewRepositoryはqueryer、これはもともとsqlxが満たしているメソッド
	//重い抽象ではない軽い抽象

	taskListUC := list.New(taskRepo)
	taskCreateUC := create.New(taskRepo, clk)
	taskUpdateUC := update.New(taskRepo, clk)
	taskDeleteUC := remove.New(taskRepo)

	taskListHandler := task.NewList(taskListUC, lg)
	taskCreateHandler := task.NewCreate(taskCreateUC, lg)
	taskUpdateHandler := task.NewUpdate(taskUpdateUC, lg)
	taskDeleteHandler := task.NewDelete(taskDeleteUC, lg)
	//user系
	var makeTxDeps txrunner.RegisterDepsFactory = func(q db.QueryerExecer) usetx.RegisterDeps {
		u := userrepo.NewRepository(q)
		v := emailverifyrepo.NewRepository(q)
		return txdeps.NewRegister(u, v)
	} //これ関数ので返す、db.QueryerExecerは別に格納しない。ここが分かりにくい
	/*func newRegisterDeps(q db.QueryerExecer) usetx.RegisterDeps {
	    u := userrepo.NewRepository(q)
	    v := emailverifyrepo.NewRepository(q)
	    return txdeps.NewRegister(u, v)
	}

	var makeTxDeps txrunner.RegisterDepsFactory = newRegisterDeps
	これを省略した記法　depsの最終DI部分*/
	beginner := xdb

	txOpts := &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	}
	//Isolation := 分離レベル、他のトランザクションがやっている途中のデータをどこまで見えるようにするか.
	//LevelReadCommittedはcommit済みのデータだけ読める。これはほぼデファクト
	//ReadOnly:=更新可能かどうか。trueは読み取り専用。動作効率の違い
	runner := txrunner.New(beginner, txOpts, makeTxDeps) //appを経由
	passwordHasher := security.NewBcryptHasher(14)
	tokenGenerator := security.NewRandomTokenGenerator(32)
	tokenHasher := security.SHA256TokenHasher{}
	registerUC := register.New( //usecaseに与える
		runner,
		clk,
		passwordHasher,
		tokenGenerator,
		tokenHasher,
	)
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
	//middlewareのチェーンはrouter/middlewareに託してもいい
	h = middleware.RequestID(h)
	// h = middleware.Recover(lg)(h)
	h = middleware.AccessLog(lg)(h)

	return h, cleanup, nil
}

//repoは永続単位->taskに対する永続はすべてtaskrepo
//ucはユーザーの行動単位->taskに対するユーザーの操作のそれぞれは分けるべき->ファイル分け
//handlerはリソースの識別単位->ファイル分け

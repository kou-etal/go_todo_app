package db

import (
	"context"
	"fmt"
	"time"

	"github.com/XSAM/otelsql" //sqlをotelsqlでwrapする。返り値は変わらない。
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/kou-etal/go_todo_app/internal/config"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewMySQL(ctx context.Context, cfg *config.Config) (*sqlx.DB, func(), error) {
	db, err := otelsql.Open("mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword,
			cfg.DBHost, cfg.DBPort,
			cfg.DBName,
		),
		otelsql.WithAttributes(semconv.DBSystemMySQL),
		//スパン-httpリクエストの中の作業単位
		/*
					[POST /tasks] 250ms                    ← スパン①（親）                                                                                     ├── [middleware.Auth] 20ms           ← スパン②
			    ├── [usecase.Create] 200ms           ← スパン③
			    │     ├── [db.Query INSERT] 150ms    ← スパン④（otelsql が自動生成）
			    │     └── [cache.Set] 25ms           ← スパン⑤
			    └── [response.JSON] 5ms             ← スパン⑥
		*/
		//スパンにMysql使ってることを伝える。GrafanaのUIで db.system = "mysql" でフィルタできる。
	)
	if err != nil {
		return nil, func() {}, err
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, func() { _ = db.Close() }, err
	}
	xdb := sqlx.NewDb(db, "mysql")

	xdb.SetMaxOpenConns(100)
	xdb.SetMaxIdleConns(100)
	xdb.SetConnMaxLifetime(5 * time.Minute)
	return xdb, func() { _ = xdb.Close() }, nil

}

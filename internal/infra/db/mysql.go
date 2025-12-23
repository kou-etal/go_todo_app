package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/kou-etal/go_todo_app/internal/config"
)

func NewMySQL(ctx context.Context, cfg *config.Config) (*sqlx.DB, func(), error) {
	db, err := sql.Open("mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?parseTime=true",
			cfg.DBUser, cfg.DBPassword,
			cfg.DBHost, cfg.DBPort,
			cfg.DBName,
		),
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
	//適当　tmp　学習用メモ、これらの定義をmainでするとmainが汚れてアンチパターン
	xdb.SetMaxOpenConns(25)
	xdb.SetMaxIdleConns(25)
	xdb.SetConnMaxLifetime(5 * time.Minute)
	return xdb, func() { _ = xdb.Close() }, nil
	//学習用メモ、main側でDBの終わりを決めたいからxdb.Close()返す
}

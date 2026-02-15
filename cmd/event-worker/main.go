package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kou-etal/go_todo_app/internal/config"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	taskeventrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/event"
	s3infra "github.com/kou-etal/go_todo_app/internal/infra/s3"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/worker/outbox"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("event-worker terminated with error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	xdb, closeDB, err := db.NewMySQL(ctx, cfg)
	if err != nil {
		return fmt.Errorf("start mysql: %w", err)
	}
	defer closeDB()

	s3Cfg := s3infra.Config{
		//os.Getenv は手作業で1個ずつ読む。caarlos0/env はstruct に自動で格納。
		Bucket:          os.Getenv("S3_BUCKET"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("S3_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		ForcePathStyle:  os.Getenv("S3_FORCE_PATH_STYLE") == "true",
	}
	uploader, err := s3infra.NewUploader(ctx, s3Cfg)
	if err != nil {
		return fmt.Errorf("init s3 uploader: %w", err)
	}

	repo := taskeventrepo.New(xdb)
	l := logger.NewSlog()
	workerCfg := outbox.DefaultConfig()

	w := outbox.NewWorker(repo, uploader, workerCfg, l)
	return w.Run(ctx)
}

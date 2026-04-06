package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/kou-etal/go_todo_app/internal/config"
	"github.com/kou-etal/go_todo_app/internal/infra/db"
	taskeventrepo "github.com/kou-etal/go_todo_app/internal/infra/repository/event"
	s3infra "github.com/kou-etal/go_todo_app/internal/infra/s3"
	"github.com/kou-etal/go_todo_app/internal/logger"
	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
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

	//defaultregistryの場合一個しかないからすべてそこに入る。
	//自前registryやからoutboxとcompactionを使う側で分岐できる。
	mp := metrics.NewProvider() //使う側のDIで分岐。
	outboxMetrics := metrics.NewOutboxMetrics(mp.Registry)
	w := outbox.NewWorker(repo, uploader, workerCfg, l, outboxMetrics)
	//w.Run()で使う。

	// metrics HTTP サーバー
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9090"
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", mp.Handler())
	metricsSrv := &http.Server{Addr: ":" + metricsPort, Handler: mux}

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return w.Run(ctx)
	})

	eg.Go(func() error {
		log.Printf("metrics server listening on :%s", metricsPort)
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("metrics server: %w", err)
		}
		return nil
	})

	eg.Go(func() error {
		<-ctx.Done()
		return metricsSrv.Close()
	})

	return eg.Wait()
}

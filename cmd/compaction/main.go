package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/push"

	s3infra "github.com/kou-etal/go_todo_app/internal/infra/s3"
	"github.com/kou-etal/go_todo_app/internal/observability/metrics"
	"github.com/kou-etal/go_todo_app/internal/worker/compaction"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("compaction terminated with error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	dateFlag := flag.String("date", "", "target date (YYYY-MM-DD). defaults to yesterday")
	flag.Parse()

	var target time.Time
	if *dateFlag == "" {
		target = time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	} else {
		parsed, err := time.Parse("2006-01-02", *dateFlag)
		if err != nil {
			return fmt.Errorf("parse date flag: %w", err)
		}
		target = parsed
	}

	s3Cfg := s3infra.Config{
		Bucket:          os.Getenv("S3_BUCKET"),
		Endpoint:        os.Getenv("S3_ENDPOINT"),
		Region:          os.Getenv("S3_REGION"),
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		ForcePathStyle:  os.Getenv("S3_FORCE_PATH_STYLE") == "true",
	}
	storage, err := s3infra.NewUploader(ctx, s3Cfg)
	if err != nil {
		return fmt.Errorf("init s3: %w", err)
	}

	cfg := compaction.DefaultConfig()
	cfg.S3Bucket = s3Cfg.Bucket
	cfg.S3Endpoint = s3Cfg.Endpoint

	mp := metrics.NewProvider()
	cm := metrics.NewCompactionMetrics(mp.Registry)
	w := compaction.NewWorker(storage, cfg, slog.Default(), cm)

	runStart := time.Now()
	runErr := w.Run(ctx, target)
	cm.LastRunDuration.Set(time.Since(runStart).Seconds())

	if runErr != nil {
		cm.LastRunStatus.Set(0)
		cm.LastFailureTimestamp.SetToCurrentTime()
	} else {
		cm.LastRunStatus.Set(1)
		cm.LastSuccessTimestamp.SetToCurrentTime()
	}

	// Pushgateway に push
	pushURL := os.Getenv("PUSHGATEWAY_URL")
	if pushURL == "" {
		pushURL = "http://pushgateway:9091"
	}
	pusher := push.New(pushURL, "compaction").Gatherer(mp.Registry)
	if err := pusher.Push(); err != nil {
		log.Printf("pushgateway push failed: %v", err)
	}

	return runErr
}

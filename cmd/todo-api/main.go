package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	app "github.com/kou-etal/go_todo_app/internal/app/httpapp"
	"github.com/kou-etal/go_todo_app/internal/app/server"
	"github.com/kou-etal/go_todo_app/internal/config"
)

// TODO:Shutdownにtimeoutがない
func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to terminated server: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	cfg, err := config.New()
	if err != nil {
		return fmt.Errorf("get config: %w", err)
	}

	apiHandler, metricsHandler, cleanup, err := app.Build(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	defer cleanup()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		return fmt.Errorf("listen port %d: %w", cfg.Port, err)
	}
	log.Printf("server started: http://%s", l.Addr().String())

	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metricsHandler)
	metricsSrv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.MetricsPort), Handler: metricsMux}

	s := server.NewServer(l, apiHandler)

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return s.Run(ctx)
	})

	eg.Go(func() error {
		log.Printf("metrics server listening on :%d", cfg.MetricsPort)
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

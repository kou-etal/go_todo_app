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

	app "github.com/kou-etal/go_todo_app/internal/app/httpapp"
	"github.com/kou-etal/go_todo_app/internal/config"
	"golang.org/x/sync/errgroup"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Printf("failed to terminated server: %v", err) //ここはloggerも使えない。
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

	h, cleanup, err := app.Build(ctx)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	defer cleanup()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))

	if err != nil {
		return fmt.Errorf("listen port %d: %w", cfg.Port, err)
	}
	log.Printf("server started: http://%s", l.Addr().String())
	srv := &http.Server{
		Handler: h,
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		if err := srv.Serve(l); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	})

	<-ctx.Done()
	log.Printf("shutdown signal received")
	if err := srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown: %w", err)
	}
	return eg.Wait()
}

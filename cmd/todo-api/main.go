package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	app "github.com/kou-etal/go_todo_app/internal/app/httpapp"
	"github.com/kou-etal/go_todo_app/internal/app/server"
	"github.com/kou-etal/go_todo_app/internal/config"
)

// TODO:Shutdownにtimeoutがない
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

	h, cleanup, err := app.Build(ctx, cfg)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	defer cleanup()

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))

	if err != nil {
		return fmt.Errorf("listen port %d: %w", cfg.Port, err)
	}
	log.Printf("server started: http://%s", l.Addr().String())
	s := server.NewServer(l, h)

	if err := s.Run(ctx); err != nil {
		return fmt.Errorf("run server: %w", err)
	}
	return nil
}

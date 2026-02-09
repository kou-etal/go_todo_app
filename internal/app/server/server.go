package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"
)

type Server struct {
	srv *http.Server
	l   net.Listener
}

func NewServer(l net.Listener, mux http.Handler) *Server {
	return &Server{
		srv: &http.Server{Handler: mux},
		l:   l,
	}
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		if err := s.srv.Serve(s.l); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve http: %w", err)
		}
		return nil
	})

	<-ctx.Done()

	if err := s.srv.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	return eg.Wait()
}

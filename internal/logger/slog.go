package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

type slogLogger struct {
	l *slog.Logger
}

func NewSlog() Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{

		Level: slog.LevelInfo, //TODO:これenvで切り替えてもいい。ローカルと運用
	})

	return &slogLogger{
		l: slog.New(handler),
	}
}

func (s *slogLogger) Debug(ctx context.Context, msg string, err error, attrs ...Attr) {
	s.l.Debug(msg, s.buildAttrs(ctx, err, attrs...)...)
}

func (s *slogLogger) Info(ctx context.Context, msg string, attrs ...Attr) {
	s.l.Info(msg, s.buildAttrs(ctx, nil, attrs...)...)
}

func (s *slogLogger) Warn(ctx context.Context, msg string, attrs ...Attr) {
	s.l.Warn(msg, s.buildAttrs(ctx, nil, attrs...)...)
}

func (s *slogLogger) Error(ctx context.Context, msg string, err error, attrs ...Attr) {
	s.l.Error(msg, s.buildAttrs(ctx, err, attrs...)...)
}

func (s *slogLogger) buildAttrs(
	ctx context.Context,
	err error,
	attrs ...Attr,
) []any {
	slogAttrs := make([]any, 0, len(attrs)+2)

	if rid, ok := requestid.FromContext(ctx); ok {
		slogAttrs = append(slogAttrs, slog.String("request_id", rid))
	}

	if err != nil {
		slogAttrs = append(slogAttrs, slog.Any("error", err))
	}

	for _, a := range attrs {
		slogAttrs = append((slogAttrs), slog.Any(a.Key, a.Value))
	}

	return slogAttrs
}

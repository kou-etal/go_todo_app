package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

// もしfromcontextヘルパーをmiddlewareで定義してたらloggerがpresentation/http/middlewareをimportし、依存逆転
// さらにgrpcでloggerとか使うときにhttpがついてくる。横断パッケージはpresentation層に依存してはならない。
type slogLogger struct {
	l *slog.Logger
}

func NewSlog() Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, //TODO:これenvで切り替えてもいい
	})

	return &slogLogger{
		l: slog.New(handler),
	}
}

func (s *slogLogger) Debug(ctx context.Context, msg string, err error, attrs ...Attr) {
	s.l.Debug(msg, s.buildAttrs(ctx, err, attrs...)...)
} //debugはエラーを返さない設計だったがエラーを返す設計に変更。
//TODO:それに伴いセキュリティを考慮しエラーをsafeエラーかどうか判定する実装必須。

func (s *slogLogger) Info(ctx context.Context, msg string, attrs ...Attr) {
	s.l.Info(msg, s.buildAttrs(ctx, nil, attrs...)...)
}

func (s *slogLogger) Error(ctx context.Context, msg string, err error, attrs ...Attr) {
	s.l.Error(msg, s.buildAttrs(ctx, err, attrs...)...)
}

func (s *slogLogger) buildAttrs(
	ctx context.Context,
	err error,
	attrs ...Attr,
) []any {
	out := make([]any, 0, len(attrs)+2)

	// request_id を context から吸い上げる
	if rid, ok := requestid.FromContext(ctx); ok {
		out = append(out, slog.String("request_id", rid))
	}

	if err != nil {
		out = append(out, slog.Any("error", err))
	}

	for _, a := range attrs {
		out = append(out, slog.Any(a.Key, a.Value))
	}

	return out
}

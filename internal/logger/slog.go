package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
	"go.opentelemetry.io/otel/trace"
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
	slogAttrs := make([]any, 0, len(attrs)+4)

	if rid, ok := requestid.FromContext(ctx); ok {
		slogAttrs = append(slogAttrs, slog.String("request_id", rid))
	}
	//requestidは自分で埋め込んだがspanidとtraceidはotel sdkは自動で埋める。便利すぎる。
	/*
	   スパンを開始するたびに自動生成
	        ctx, span := otel.Tracer("usecase").Start(ctx, "Create")
	*/
	//trace.SpanContextFromContext(ctx);は返り値一つ。
	if spanid := trace.SpanContextFromContext(ctx); spanid.IsValid() {
		//SpanContextFromContext(ctx)は読み取り専用。SpanFromContext → Span を操作したい場合
		//スパンが作られてない ctx を受け取った場合に意味ない trace_id をログに出さない。spanCtx.IsValid
		/*
					スパンあり: trace_id=abc123, span_id=def456 → ログに出す
			  スパンなし: trace_id=000000, span_id=000000 → 出さない
			  例えばヘルスチェックやバッチ処理でスパンを作ってない ctx あり得る。
		*/
		slogAttrs = append(slogAttrs, slog.String("trace_id", spanid.TraceID().String()))
		slogAttrs = append(slogAttrs, slog.String("span_id", spanid.SpanID().String()))
	}

	if err != nil {
		slogAttrs = append(slogAttrs, slog.Any("error", err))
	}

	for _, a := range attrs {
		slogAttrs = append((slogAttrs), slog.Any(a.Key, a.Value))
	}

	return slogAttrs
}

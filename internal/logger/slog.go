package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/kou-etal/go_todo_app/internal/observability/requestid"
)

//logger抽象を作り、最終は具体slogのdebugを使う。

// もしfromcontextヘルパーをmiddlewareで定義してたらloggerがpresentation/http/middlewareをimportし、依存逆転
// さらにgrpcでloggerとか使うときにhttpがついてくる。横断パッケージはpresentation層に依存してはならない。
type slogLogger struct {
	l *slog.Logger
}

func NewSlog() Logger { //ログはstream形式。stdoutで出力。metricsは/metricsでエンドポイント作ってprometheusとかで取りに行く
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{ //logをjsonで返す。開発環境はtxt。運用はjsonが良くある
		//どこまでログに出すか
		//実際こういうテンプレはドキュメント+AIの併用するべき。
		Level: slog.LevelInfo, //TODO:これenvで切り替えてもいい。ローカルと運用
	})

	return &slogLogger{
		l: slog.New(handler),
	}
}

//attrsにはログを後から調査するときに必要なデータを入れる。user_id,task_id
//ここをちゃんと再現できるようにしないと使い物にならないログ
//機密は出力ない

func (s *slogLogger) Debug(ctx context.Context, msg string, err error, attrs ...Attr) {
	s.l.Debug(msg, s.buildAttrs(ctx, err, attrs...)...) //表はlogger。具体はslog
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
	attrs ...Attr, //これはnullでもいい
) []any {
	slogAttrs := make([]any, 0, len(attrs)+2)

	// request_id を context から吸い上げる
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

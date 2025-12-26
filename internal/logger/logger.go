package logger

import (
	"context"
)

/* slog /zap/zerologを差し替えるためにinterface
interfaceは使う側で定義しろってのは使う側が必要なメソッドだけ用意してむだにinterface作るなって意味。
これは横断的やからここにおいていい。
例えばusecase側で
type usecaseLogger interface {
  Info(ctx context.Context, msg string, attrs ...logger.Attr)
  Error(ctx context.Context, msg string, err error, attrs ...logger.Attr)
}絞ったinterface*/

type Logger interface {
	Debug(ctx context.Context, msg string, err error, attrs ...Attr)
	Info(ctx context.Context, msg string, attrs ...Attr)
	Error(ctx context.Context, msg string, err error, attrs ...Attr)
}

// slog.Attrとかを消すためにAttr
type Attr struct {
	Key   string
	Value any
}

func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

//Attr{Key: key, Value: value}って毎回記述するのだるい
//これも結局ヘルパーはutilsで切ってない

func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

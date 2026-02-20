package logger

import (
	"context"
)

type Logger interface {
	Debug(ctx context.Context, msg string, err error, attrs ...Attr)
	Info(ctx context.Context, msg string, attrs ...Attr)
	Warn(ctx context.Context, msg string, attrs ...Attr)
	Error(ctx context.Context, msg string, err error, attrs ...Attr)
}

type Attr struct {
	Key   string
	Value any
}

func String(key, value string) Attr {
	return Attr{Key: key, Value: value}
}

func Int(key string, value int) Attr {
	return Attr{Key: key, Value: value}
}

func Int64(key string, value int64) Attr {
	return Attr{Key: key, Value: value}
}

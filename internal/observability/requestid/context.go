package requestid

import "context"

type ctxKeyRequestID struct{}

var requestIDKey = ctxKeyRequestID{}

func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

func WithContext(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

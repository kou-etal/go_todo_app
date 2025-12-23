package requestid

import "context"

// ctxKeyRequestID は衝突を避けるため。string、type ctxKeyRequestID stringは不可
type ctxKeyRequestID struct{}

var requestIDKey = ctxKeyRequestID{}

// FromContext は logger 等から request_id を取得するための helper
// ここはerrで返してない。つまりあるか確認するだけ。
// middlewareで保証してるけどここでもerrで保険は良くない。
// べつにidがなくてもロジックには問題ないゆえに保険かけない。
//仮にrequestidがロジックに入ってるならdomainVOとか別の境界で保証する保険あってもいい。
func FromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey).(string)
	return id, ok
}

// WithContext は request_id を context に埋め込むための helper。
// 重要：context key の所有権をこのパッケージに寄せるため、middleware 側で WithValue 直呼びしない。
func WithContext(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

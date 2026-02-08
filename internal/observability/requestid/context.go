package requestid

import "context"

// ctxKeyRequestID は衝突を避けるため。string、type ctxKeyRequestID stringは不可
//もっと詳しく。まずはstringは普通に衝突する
//type Mykey stringは別パッケージからも作れる
//type mykey stringは安全やけどmykey("request_id")のrequest_id部分に意味があるように思える。mykey("trace_id")とかで使ってしまう可能性
//type ctxKeyRequestID struct{}は衝突しにくい。明確にrequest_id用途だけ。
//keyはstruct
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
//context key の所有権をこのパッケージに寄せるため、middleware 側で WithValue 直呼びしない。
func WithContext(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestIDKey, rid)
}

//log、metrics、traceで観測。accesslogは観測ではあるけど別。

package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func NewProvider(ctx context.Context, serviceName, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	//送る場所定義
	//Goはかなり細かく定義する。TSはデフォルトの設定が多い。
	//otel sdkは実際は環境変数自動で読む。 exporter, err := otlptracehttp.New(ctx)で動く。
	//今回はブラックボックスなくすためにconfigで扱う。
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
		//TLSなし。ローカル用。dockerでのコンテナ通信は外部に出ないためTLSなし。デプロイ環境ではTLS。
		//TODO:ハードコードせずにWithInsecureは環境変数で切り替える。
	)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}

	//サービス名を記述。 全スパンに付与されるメタデータ。バックエンドの UI でサービス名ごとにトレースを絞り込める。
	//semconv = Semantic Conventions。 OTel が定めた属性名の標準規約
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName), //go_todo_app
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel resource: %w", err)
	}
	//TracerProvider assemble
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // バッチで送信（1件ずつではなくまとめる）
		sdktrace.WithResource(res),
	)
	// グローバルに登録。これにより、アプリのどこからでも otel.Tracer("") で使えるようになる。
	// otel.Tracer("usecase").Start(ctx, "Create")
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	// propagation.TraceContext{} 標準形式。
	// SetTextMapPropagator  HTTPヘッダでトレースを伝播する方式を設定する
	/*
			設定あり:
		    [handler] ──HTTP──► [外部API]
		    traceparent: 00-abc123-def456-01  ← 自動でヘッダに入る
		    → 外部API側でも同じ trace_id で繋がる

		  設定なし:
		    [handler] ──HTTP──► [外部API]
		    （ヘッダに何も入らない）
		    → 外部API側では別のトレースになる。繋がらない。

	*/

	return tp, nil
}

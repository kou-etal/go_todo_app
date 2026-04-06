package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// //prometheys client golangはsdk提供してない。ゆえに自分で struct 作って Handler() や Registry()を定義する。
type Provider struct {
	Registry *prometheus.Registry
}

// 今回はoutbox compactionのメトリクスを分離するためにDefault Registry 使わずに自分でregistry作る。
func NewProvider() *Provider { //metricsは分岐なし。引数なし。共有する。
	reg := prometheus.NewRegistry()

	reg.MustRegister(collectors.NewGoCollector())
	// goroutine 数、GC 回数、メモリ使用量など go_* 系
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	//CPU 使用時間、オープン FD 数、メモリ RSS など process_* 系
	//これらはDefaultRegisterer使うと自動で定義される。
	return &Provider{Registry: reg}
}

func (p *Provider) Handler() http.Handler {
	return promhttp.HandlerFor(p.Registry, promhttp.HandlerOpts{})
} // /metricsでRegistry に登録されてる全メトリクスを返すハンドラー。
//使う側はメトリクスを登録することで分岐する。

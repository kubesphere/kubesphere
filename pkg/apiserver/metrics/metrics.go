/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package metrics

import (
	"sync"

	"github.com/emicklei/go-restful/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
	componentbasemetrics "k8s.io/component-base/metrics"

	ksVersion "kubesphere.io/kubesphere/pkg/version"
)

var (
	registerOnce sync.Once

	Registry = componentbasemetrics.NewKubeRegistry()

	RequestCounter = componentbasemetrics.NewCounterVec(
		&componentbasemetrics.CounterOpts{
			Name:           "ks_server_request_total",
			Help:           "Counter of ks_server requests broken out for each verb, group, version, resource and HTTP response code.",
			StabilityLevel: componentbasemetrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource", "code"},
	)

	RequestLatencies = componentbasemetrics.NewHistogramVec(
		&componentbasemetrics.HistogramOpts{
			Name: "ks_server_request_duration_seconds",
			Help: "Response latency distribution in seconds for each verb, group, version, resource",
			// This metric is used for verifying api call latencies SLO,
			// as well as tracking regressions in this aspects.
			// Thus we customize buckets significantly, to empower both usecases.
			Buckets: []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
				1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
			StabilityLevel: componentbasemetrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource"},
	)

	metricsList = []componentbasemetrics.Registerable{
		RequestCounter,
		RequestLatencies,
	}
)

func init() {
	componentbasemetrics.BuildVersion = versionGet
}

func versionGet() apimachineryversion.Info {
	info := ksVersion.Get()
	return apimachineryversion.Info{
		Major:        info.GitMajor,
		Minor:        info.GitMinor,
		GitVersion:   info.GitVersion,
		GitCommit:    info.GitCommit,
		GitTreeState: info.GitTreeState,
		BuildDate:    info.BuildDate,
		GoVersion:    info.GoVersion,
		Compiler:     info.Compiler,
		Platform:     info.Platform,
	}
}

func registerMetrics() {
	Registry.Registerer().MustRegister(collectors.NewGoCollector())
	Registry.Registerer().MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	for _, m := range metricsList {
		Registry.MustRegister(m)
	}
}

func Install(c *restful.Container) {
	registerOnce.Do(registerMetrics)
	c.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(prometheus.NewRegistry(), promhttp.HandlerFor(Registry, promhttp.HandlerOpts{})),
	)
}

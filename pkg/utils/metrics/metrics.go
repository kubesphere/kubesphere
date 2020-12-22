package metrics

import (
	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
	compbasemetrics "k8s.io/component-base/metrics"
	ksVersion "kubesphere.io/kubesphere/pkg/version"
	"net/http"
	"sync"
)

var (
	Defaults DefaultMetrics
	//registerMetrics sync.Once
	defaultRegistry = compbasemetrics.NewKubeRegistry()
	// MustRegister registers registerable metrics but uses the defaultRegistry, panic upon the first registration that causes an error
	MustRegister = defaultRegistry.MustRegister
	// Register registers a collectable metric but uses the defaultRegistry
	Register = defaultRegistry.Register
)

// DefaultMetrics installs the default prometheus metrics handler
type DefaultMetrics struct{}

// Install adds the DefaultMetrics handler
func (m DefaultMetrics) Install(c *restful.Container) {
	register()
	c.Handle("/kapis/metrics", Handler())
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

// Handler returns an HTTP handler for the DefaultGatherer. It is
// already instrumented with InstrumentHandler (using "prometheus" as handler
// name).
func Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{}))
}

var registerMetrics sync.Once

func register() {
	registerMetrics.Do(func() {
		defaultRegistry.RawMustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
		defaultRegistry.RawMustRegister(prometheus.NewGoCollector())
		for _, metric := range metrics {
			MustRegister(metric)
		}
	})
}

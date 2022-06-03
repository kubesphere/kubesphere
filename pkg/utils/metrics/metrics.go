// /*
// Copyright 2020 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */
//

package metrics

import (
	"net/http"
	"sync"

	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	apimachineryversion "k8s.io/apimachinery/pkg/version"
	compbasemetrics "k8s.io/component-base/metrics"

	ksVersion "kubesphere.io/kubesphere/pkg/version"
)

var (
	registerOnce sync.Once

	Defaults        DefaultMetrics
	defaultRegistry compbasemetrics.KubeRegistry
	// MustRegister registers registerable metrics but uses the defaultRegistry, panic upon the first registration that causes an error
	MustRegister func(...compbasemetrics.Registerable)
	// Register registers a collectable metric but uses the defaultRegistry
	Register func(compbasemetrics.Registerable) error

	RawMustRegister func(...prometheus.Collector)
)

func init() {
	compbasemetrics.BuildVersion = versionGet

	defaultRegistry = compbasemetrics.NewKubeRegistry()
	MustRegister = defaultRegistry.MustRegister
	Register = defaultRegistry.Register
	RawMustRegister = defaultRegistry.RawMustRegister
}

// DefaultMetrics installs the default prometheus metrics handler
type DefaultMetrics struct{}

// Install adds the DefaultMetrics handler
func (m DefaultMetrics) Install(c *restful.Container) {
	registerOnce.Do(m.registerMetrics)
	c.Handle("/kapis/metrics", Handler())
}

func (m DefaultMetrics) registerMetrics() {
	RawMustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	RawMustRegister(prometheus.NewGoCollector())
}

//Overwrite version.Get
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
	return promhttp.InstrumentMetricHandler(prometheus.NewRegistry(), promhttp.HandlerFor(defaultRegistry, promhttp.HandlerOpts{}))
}

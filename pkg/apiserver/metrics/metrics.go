// Copyright 2023 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func registerBaseMetrics() {
	Registry.Registerer().MustRegister(collectors.NewGoCollector())
	Registry.Registerer().MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
}

func Install(c *restful.Container) {
	registerOnce.Do(registerBaseMetrics)
	c.Handle(
		"/metrics",
		promhttp.InstrumentMetricHandler(prometheus.NewRegistry(), promhttp.HandlerFor(Registry, promhttp.HandlerOpts{})),
	)
}

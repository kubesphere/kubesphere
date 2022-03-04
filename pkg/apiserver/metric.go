// Copyright 2022 The KubeSphere Authors.
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
//
package apiserver

import (
	compbasemetrics "k8s.io/component-base/metrics"

	"kubesphere.io/kubesphere/pkg/utils/metrics"
)

var (
	RequestCounter = compbasemetrics.NewCounterVec(
		&compbasemetrics.CounterOpts{
			Name:           "ks_server_request_total",
			Help:           "Counter of ks_server requests broken out for each verb, group, version, resource and HTTP response code.",
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource", "code"},
	)

	RequestLatencies = compbasemetrics.NewHistogramVec(
		&compbasemetrics.HistogramOpts{
			Name: "ks_server_request_duration_seconds",
			Help: "Response latency distribution in seconds for each verb, group, version, resource",
			// This metric is used for verifying api call latencies SLO,
			// as well as tracking regressions in this aspects.
			// Thus we customize buckets significantly, to empower both usecases.
			Buckets: []float64{0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
				1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
			StabilityLevel: compbasemetrics.ALPHA,
		},
		[]string{"verb", "group", "version", "resource"},
	)

	metricsList = []compbasemetrics.Registerable{
		RequestCounter,
		RequestLatencies,
	}
)

func registerMetrics() {
	for _, m := range metricsList {
		metrics.MustRegister(m)
	}
}

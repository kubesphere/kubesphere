/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/client-go/util/workqueue"
	logf "sigs.k8s.io/controller-runtime/pkg/internal/log"
)

var log = logf.RuntimeLog.WithName("metrics")

// This file is copied and adapted from k8s.io/kubernetes/pkg/util/workqueue/prometheus
// which registers metrics to the default prometheus Registry. We require very
// similar functionality, but must register metrics to a different Registry.

func init() {
	workqueue.SetProvider(workqueueMetricsProvider{})
}

func registerWorkqueueMetric(c prometheus.Collector, name, queue string) {
	if err := Registry.Register(c); err != nil {
		log.Error(err, "failed to register metric", "name", name, "queue", queue)
	}
}

type workqueueMetricsProvider struct{}

func (workqueueMetricsProvider) NewDepthMetric(queue string) workqueue.GaugeMetric {
	const name = "workqueue_depth"
	m := prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        name,
		Help:        "Current depth of workqueue",
		ConstLabels: prometheus.Labels{"name": queue},
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewAddsMetric(queue string) workqueue.CounterMetric {
	const name = "workqueue_adds_total"
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        "Total number of adds handled by workqueue",
		ConstLabels: prometheus.Labels{"name": queue},
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewLatencyMetric(queue string) workqueue.HistogramMetric {
	const name = "workqueue_queue_duration_seconds"
	m := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        name,
		Help:        "How long in seconds an item stays in workqueue before being requested.",
		ConstLabels: prometheus.Labels{"name": queue},
		Buckets:     prometheus.ExponentialBuckets(10e-9, 10, 10),
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewWorkDurationMetric(queue string) workqueue.HistogramMetric {
	const name = "workqueue_work_duration_seconds"
	m := prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:        name,
		Help:        "How long in seconds processing an item from workqueue takes.",
		ConstLabels: prometheus.Labels{"name": queue},
		Buckets:     prometheus.ExponentialBuckets(10e-9, 10, 10),
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewUnfinishedWorkSecondsMetric(queue string) workqueue.SettableGaugeMetric {
	const name = "workqueue_unfinished_work_seconds"
	m := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: "How many seconds of work has done that " +
			"is in progress and hasn't been observed by work_duration. Large " +
			"values indicate stuck threads. One can deduce the number of stuck " +
			"threads by observing the rate at which this increases.",
		ConstLabels: prometheus.Labels{"name": queue},
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewLongestRunningProcessorSecondsMetric(queue string) workqueue.SettableGaugeMetric {
	const name = "workqueue_longest_running_processor_seconds"
	m := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: name,
		Help: "How many seconds has the longest running " +
			"processor for workqueue been running.",
		ConstLabels: prometheus.Labels{"name": queue},
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

func (workqueueMetricsProvider) NewRetriesMetric(queue string) workqueue.CounterMetric {
	const name = "workqueue_retries_total"
	m := prometheus.NewCounter(prometheus.CounterOpts{
		Name:        name,
		Help:        "Total number of retries handled by workqueue",
		ConstLabels: prometheus.Labels{"name": queue},
	})
	registerWorkqueueMetric(m, name, queue)
	return m
}

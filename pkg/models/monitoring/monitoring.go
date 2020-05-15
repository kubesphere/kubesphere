/*

 Copyright 2019 The KubeSphere Authors.

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

package monitoring

import (
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/monitoring/expressions"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"time"
)

type MonitoringOperator interface {
	GetMetric(expr, namespace string, time time.Time) (monitoring.Metric, error)
	GetMetricOverTime(expr, namespace string, start, end time.Time, step time.Duration) (monitoring.Metric, error)
	GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics
	GetMetadata(namespace string) Metadata
	GetMetricLabels(metric, namespace string, start, end time.Time) monitoring.MetricLabels
}

type monitoringOperator struct {
	c monitoring.Interface
}

func NewMonitoringOperator(client monitoring.Interface) MonitoringOperator {
	return &monitoringOperator{client}
}

func (mo monitoringOperator) GetMetric(expr, namespace string, time time.Time) (monitoring.Metric, error) {
	// Different monitoring backend implementations have different ways to enforce namespace isolation.
	// Each implementation should register itself to `ReplaceNamespaceFns` during init().
	// We hard code "prometheus" here because we only support this datasource so far.
	// In the future, maybe the value should be returned from a method like `mo.c.GetMonitoringServiceName()`.
	expr, err := expressions.ReplaceNamespaceFns["prometheus"](expr, namespace)
	if err != nil {
		return monitoring.Metric{}, err
	}
	return mo.c.GetMetric(expr, time), nil
}

func (mo monitoringOperator) GetMetricOverTime(expr, namespace string, start, end time.Time, step time.Duration) (monitoring.Metric, error) {
	// Different monitoring backend implementations have different ways to enforce namespace isolation.
	// Each implementation should register itself to `ReplaceNamespaceFns` during init().
	// We hard code "prometheus" here because we only support this datasource so far.
	// In the future, maybe the value should be returned from a method like `mo.c.GetMonitoringServiceName()`.
	expr, err := expressions.ReplaceNamespaceFns["prometheus"](expr, namespace)
	if err != nil {
		return monitoring.Metric{}, err
	}
	return mo.c.GetMetricOverTime(expr, start, end, step), nil
}

func (mo monitoringOperator) GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics {
	ress := mo.c.GetNamedMetrics(metrics, time, opt)
	return Metrics{Results: ress}
}

func (mo monitoringOperator) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics {
	ress := mo.c.GetNamedMetricsOverTime(metrics, start, end, step, opt)
	return Metrics{Results: ress}
}

func (mo monitoringOperator) GetMetadata(namespace string) Metadata {
	data := mo.c.GetMetadata(namespace)
	return Metadata{Data: data}
}

func (mo monitoringOperator) GetMetricLabels(metric, namespace string, start, end time.Time) monitoring.MetricLabels {
	var permit bool

	availMetrics := mo.c.GetMetadata(namespace)
	for _, item := range availMetrics {
		if item.Metric == metric {
			permit = true
			break
		}
	}

	if !permit {
		return monitoring.MetricLabels{}
	}

	// Different monitoring backend implementations have different ways to enforce namespace isolation.
	// Each implementation should register itself to `ReplaceNamespaceFns` during init().
	// We hard code "prometheus" here because we only support this datasource so far.
	// In the future, maybe the value should be returned from a method like `mo.c.GetMonitoringServiceName()`.
	expr, err := expressions.ReplaceNamespaceFns["prometheus"](metric, namespace)
	if err != nil {
		klog.Error(err)
		return monitoring.MetricLabels{}
	}
	return mo.c.GetMetricLabels(expr, start, end)
}

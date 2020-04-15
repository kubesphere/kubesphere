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
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"time"
)

type MonitoringOperator interface {
	GetMetrics(stmts []string, time time.Time) Metrics
	GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) Metrics
	GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics
}

type monitoringOperator struct {
	c monitoring.Interface
}

func NewMonitoringOperator(client monitoring.Interface) MonitoringOperator {
	return &monitoringOperator{client}
}

// TODO(huanggze): reserve for custom monitoring
func (mo monitoringOperator) GetMetrics(stmts []string, time time.Time) Metrics {
	panic("implement me")
}

// TODO(huanggze): reserve for custom monitoring
func (mo monitoringOperator) GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) Metrics {
	panic("implement me")
}

func (mo monitoringOperator) GetNamedMetrics(metrics []string, time time.Time, opt monitoring.QueryOption) Metrics {
	ress := mo.c.GetNamedMetrics(metrics, time, opt)
	return Metrics{Results: ress}
}

func (mo monitoringOperator) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt monitoring.QueryOption) Metrics {
	ress := mo.c.GetNamedMetricsOverTime(metrics, start, end, step, opt)
	return Metrics{Results: ress}
}

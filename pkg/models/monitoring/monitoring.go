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
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"time"
)

type MonitoringOperator interface {
	GetMetrics(stmts []string, time time.Time) (v1alpha2.APIResponse, error)
	GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) (v1alpha2.APIResponse, error)
	GetNamedMetrics(time time.Time, opt monitoring.QueryOption) (v1alpha2.APIResponse, error)
	GetNamedMetricsOverTime(start, end time.Time, step time.Duration, opt monitoring.QueryOption) (v1alpha2.APIResponse, error)
	SortMetrics(raw v1alpha2.APIResponse, target, order, identifier string) (v1alpha2.APIResponse, int)
	PageMetrics(raw v1alpha2.APIResponse, page, limit, rows int) v1alpha2.APIResponse
}

type monitoringOperator struct {
	c monitoring.Interface
}

func NewMonitoringOperator(client monitoring.Interface) MonitoringOperator {
	return &monitoringOperator{client}
}

// TODO(huanggze): reserve for custom monitoring
func (mo monitoringOperator) GetMetrics(stmts []string, time time.Time) (v1alpha2.APIResponse, error) {
	panic("implement me")
}

// TODO(huanggze): reserve for custom monitoring
func (mo monitoringOperator) GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) (v1alpha2.APIResponse, error) {
	panic("implement me")
}

func (mo monitoringOperator) GetNamedMetrics(time time.Time, opt monitoring.QueryOption) (v1alpha2.APIResponse, error) {
	metrics, err := mo.c.GetNamedMetrics(time, opt)
	if err != nil {
		klog.Error(err)
	}
	return v1alpha2.APIResponse{Results: metrics}, err
}

func (mo monitoringOperator) GetNamedMetricsOverTime(start, end time.Time, step time.Duration, opt monitoring.QueryOption) (v1alpha2.APIResponse, error) {
	metrics, err := mo.c.GetNamedMetricsOverTime(start, end, step, opt)
	if err != nil {
		klog.Error(err)
	}
	return v1alpha2.APIResponse{Results: metrics}, err
}

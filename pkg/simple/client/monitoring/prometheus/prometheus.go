/*
Copyright 2020 KubeSphere Authors

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

package prometheus

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
)

const MeteringDefaultTimeout = 20 * time.Second

// prometheus implements monitoring interface backed by Prometheus
type prometheus struct {
	client apiv1.API
}

func NewPrometheus(options *Options) (monitoring.Interface, error) {
	cfg := api.Config{
		Address: options.Endpoint,
	}

	client, err := api.NewClient(cfg)
	return prometheus{client: apiv1.NewAPI(client)}, err
}

func (p prometheus) GetMetric(expr string, ts time.Time) monitoring.Metric {
	var parsedResp monitoring.Metric

	value, _, err := p.client.Query(context.Background(), expr, ts)
	if err != nil {
		parsedResp.Error = err.Error()
	} else {
		parsedResp.MetricData = parseQueryResp(value, nil)
	}

	return parsedResp
}

func (p prometheus) GetMetricOverTime(expr string, start, end time.Time, step time.Duration) monitoring.Metric {
	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	value, _, err := p.client.QueryRange(context.Background(), expr, timeRange)

	var parsedResp monitoring.Metric
	if err != nil {
		parsedResp.Error = err.Error()
	} else {
		parsedResp.MetricData = parseQueryRangeResp(value, nil)
	}
	return parsedResp
}

func (p prometheus) GetNamedMetrics(metrics []string, ts time.Time, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var mtx sync.Mutex
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	for _, metric := range metrics {
		wg.Add(1)
		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}

			value, _, err := p.client.Query(context.Background(), makeExpr(metric, *opts), ts)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryResp(value, genMetricFilter(o))
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(metric)
	}

	wg.Wait()

	return res
}

func (p prometheus) GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, o monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var mtx sync.Mutex
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	for _, metric := range metrics {
		wg.Add(1)
		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}

			value, _, err := p.client.QueryRange(context.Background(), makeExpr(metric, *opts), timeRange)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryRangeResp(value, genMetricFilter(o))
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(metric)
	}

	wg.Wait()

	return res
}

func (p prometheus) GetNamedMeters(meters []string, ts time.Time, opts []monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var wg sync.WaitGroup
	var mtx sync.Mutex

	queryOptions := monitoring.NewQueryOptions()

	for _, opt := range opts {
		opt.Apply(queryOptions)
	}

	prometheusCtx, cancel := context.WithTimeout(context.Background(), MeteringDefaultTimeout)
	defer cancel()

	for _, meter := range meters {

		wg.Add(1)

		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}

			begin := time.Now()
			value, _, err := p.client.Query(prometheusCtx, makeMeterExpr(metric, *queryOptions), ts)
			end := time.Now()
			timeElapsed := end.Unix() - begin.Unix()
			if timeElapsed > int64(MeteringDefaultTimeout.Seconds())/2 {
				klog.Warningf("long time query[cost %v seconds], expr: %v", timeElapsed, makeMeterExpr(metric, *queryOptions))
			}

			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryResp(value, nil)
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(meter)

	}

	wg.Wait()

	return res
}

func (p prometheus) GetNamedMetersOverTime(meters []string, start, end time.Time, step time.Duration, opts []monitoring.QueryOption) []monitoring.Metric {
	var res []monitoring.Metric
	var wg sync.WaitGroup
	var mtx sync.Mutex

	queryOptions := monitoring.NewQueryOptions()

	for _, opt := range opts {
		opt.Apply(queryOptions)
	}

	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	prometheusCtx, cancel := context.WithTimeout(context.Background(), MeteringDefaultTimeout)
	defer cancel()

	for _, meter := range meters {

		wg.Add(1)

		go func(metric string) {
			parsedResp := monitoring.Metric{MetricName: metric}
			begin := time.Now()
			value, _, err := p.client.QueryRange(prometheusCtx, makeMeterExpr(metric, *queryOptions), timeRange)
			end := time.Now()
			timeElapsed := end.Unix() - begin.Unix()
			if timeElapsed > int64(MeteringDefaultTimeout.Seconds())/2 {
				klog.Warningf("long time query[cost %v seconds], expr: %v", timeElapsed, makeMeterExpr(metric, *queryOptions))
			}

			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryRangeResp(value, nil)
			}

			mtx.Lock()
			res = append(res, parsedResp)
			mtx.Unlock()

			wg.Done()
		}(meter)
	}

	wg.Wait()

	return res
}

func (p prometheus) GetMetadata(namespace string) []monitoring.Metadata {
	var meta []monitoring.Metadata
	var matchTarget string

	if namespace != "" {
		// Filter metrics available to members of this namespace
		matchTarget = fmt.Sprintf("{namespace=\"%s\"}", namespace)
	}
	items, err := p.client.TargetsMetadata(context.Background(), matchTarget, "", "")
	if err != nil {
		klog.Error(err)
		return meta
	}

	// Deduplication
	set := make(map[string]bool)
	for _, item := range items {
		_, ok := set[item.Metric]
		if !ok {
			set[item.Metric] = true
			meta = append(meta, monitoring.Metadata{
				Metric: item.Metric,
				Type:   string(item.Type),
				Help:   item.Help,
			})
		}
	}

	return meta
}

func (p prometheus) GetMetricLabelSet(expr string, start, end time.Time) []map[string]string {
	var res []map[string]string

	labelSet, _, err := p.client.Series(context.Background(), []string{expr}, start, end)
	if err != nil {
		klog.Error(err)
		return []map[string]string{}
	}

	for _, item := range labelSet {
		var tmp = map[string]string{}
		for key, val := range item {
			if key == "__name__" {
				continue
			}
			tmp[string(key)] = string(val)
		}

		res = append(res, tmp)
	}

	return res
}

func parseQueryRangeResp(value model.Value, metricFilter func(metric model.Metric) bool) monitoring.MetricData {
	res := monitoring.MetricData{MetricType: monitoring.MetricTypeMatrix}

	data, _ := value.(model.Matrix)

	for _, v := range data {
		if metricFilter != nil && !metricFilter(v.Metric) {
			continue
		}
		mv := monitoring.MetricValue{
			Metadata: make(map[string]string),
		}

		for k, v := range v.Metric {
			mv.Metadata[string(k)] = string(v)
		}

		for _, k := range v.Values {
			mv.Series = append(mv.Series, monitoring.Point{float64(k.Timestamp) / 1000, float64(k.Value)})
		}

		res.MetricValues = append(res.MetricValues, mv)
	}

	return res
}

func parseQueryResp(value model.Value, metricFilter func(metric model.Metric) bool) monitoring.MetricData {
	res := monitoring.MetricData{MetricType: monitoring.MetricTypeVector}

	data, _ := value.(model.Vector)

	for _, v := range data {
		if metricFilter != nil && !metricFilter(v.Metric) {
			continue
		}
		mv := monitoring.MetricValue{
			Metadata: make(map[string]string),
		}

		for k, v := range v.Metric {
			mv.Metadata[string(k)] = string(v)
		}

		mv.Sample = &monitoring.Point{float64(v.Timestamp) / 1000, float64(v.Value)}

		res.MetricValues = append(res.MetricValues, mv)
	}

	return res
}

func genMetricFilter(o monitoring.QueryOption) func(metric model.Metric) bool {
	if o != nil {
		if po, ok := o.(monitoring.PodOption); ok {
			if po.NamespacedResourcesFilter != "" {
				namespacedPodsMap := make(map[string]struct{})
				for _, s := range strings.Split(po.NamespacedResourcesFilter, "|") {
					namespacedPodsMap[s] = struct{}{}
				}
				return func(metric model.Metric) bool {
					if len(metric) == 0 {
						return false
					}
					_, ok := namespacedPodsMap[string(metric["namespace"])+"/"+string(metric["pod"])]
					return ok
				}
			}
		}
	}
	return func(metric model.Metric) bool {
		return true
	}
}

package prometheus

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"sort"
	"sync"
	"time"
)

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

	value, err := p.client.Query(context.Background(), expr, ts)
	if err != nil {
		parsedResp.Error = err.Error()
	} else {
		parsedResp.MetricData = parseQueryResp(value)
	}

	return parsedResp
}

func (p prometheus) GetMetricOverTime(expr string, start, end time.Time, step time.Duration) monitoring.Metric {
	timeRange := apiv1.Range{
		Start: start,
		End:   end,
		Step:  step,
	}

	value, err := p.client.QueryRange(context.Background(), expr, timeRange)

	var parsedResp monitoring.Metric
	if err != nil {
		parsedResp.Error = err.Error()
	} else {
		parsedResp.MetricData = parseQueryRangeResp(value)
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

			value, err := p.client.Query(context.Background(), makeExpr(metric, *opts), ts)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryResp(value)
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

			value, err := p.client.QueryRange(context.Background(), makeExpr(metric, *opts), timeRange)
			if err != nil {
				parsedResp.Error = err.Error()
			} else {
				parsedResp.MetricData = parseQueryRangeResp(value)
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

func (p prometheus) GetMetadata(namespace string) []monitoring.Metadata {
	var meta []monitoring.Metadata

	// Filter metrics available to members of this namespace
	matchTarget := fmt.Sprintf("{namespace=\"%s\"}", namespace)
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

func (p prometheus) GetMetricLabels(expr string, start, end time.Time) monitoring.MetricLabels {
	var res = make(map[string][]string)

	labelSet, err := p.client.Series(context.Background(), []string{expr}, start, end)
	if err != nil {
		klog.Error(err)
		return res
	}

	for _, label := range labelSet {
		for key, value := range label {
			if key == "__name__" {
				continue
			}

			res[string(key)] = append(res[string(key)], string(value))
		}
	}

	// Deduplicate and Sort
	for label, values := range res {
		res[label] = stringutils.Unique(values)
		sort.StringSlice(res[label]).Sort()
	}

	return res
}

func parseQueryRangeResp(value model.Value) monitoring.MetricData {
	res := monitoring.MetricData{MetricType: monitoring.MetricTypeMatrix}

	data, _ := value.(model.Matrix)

	for _, v := range data {
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

func parseQueryResp(value model.Value) monitoring.MetricData {
	res := monitoring.MetricData{MetricType: monitoring.MetricTypeVector}

	data, _ := value.(model.Vector)

	for _, v := range data {
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

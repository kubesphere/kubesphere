package prometheus

import (
	"context"
	"github.com/prometheus/client_golang/api"
	apiv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
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

// TODO(huanggze): reserve for custom monitoring
func (p prometheus) GetMetrics(stmts []string, time time.Time) []monitoring.Metric {
	panic("implement me")
}

// TODO(huanggze): reserve for custom monitoring
func (p prometheus) GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) []monitoring.Metric {
	panic("implement me")
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
				parsedResp.Error = err.(*apiv1.Error).Msg
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
				parsedResp.Error = err.(*apiv1.Error).Msg
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

		mv.Sample = monitoring.Point{float64(v.Timestamp) / 1000, float64(v.Value)}

		res.MetricValues = append(res.MetricValues, mv)
	}

	return res
}

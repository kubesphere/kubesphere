package prometheus

import (
	"fmt"
	"github.com/json-iterator/go"
	"io/ioutil"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"net/http"
	"net/url"
	"regexp"
	"sync"
	"time"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// prometheus implements monitoring interface backed by Prometheus
type prometheus struct {
	options *Options
	client  *http.Client
}

func NewPrometheus(options *Options) monitoring.Interface {
	return &prometheus{
		options: options,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// TODO(huanggze): reserve for custom monitoring
func (p *prometheus) GetMetrics(stmts []string, time time.Time) ([]monitoring.Metric, error) {
	panic("implement me")
}

// TODO(huanggze): reserve for custom monitoring
func (p *prometheus) GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) ([]monitoring.Metric, error) {
	panic("implement me")
}

func (p *prometheus) GetNamedMetrics(ts time.Time, o monitoring.QueryOption) ([]monitoring.Metric, error) {
	metrics := make([]monitoring.Metric, 0)
	var mtx sync.Mutex // guard metrics
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	errCh := make(chan error)
	for _, metric := range opts.NamedMetrics {
		matched, _ := regexp.MatchString(opts.MetricFilter, metric)
		if matched {
			exp := makeExpression(metric, *opts)
			wg.Add(1)
			go func(metric, exp string) {
				res, err := p.query(exp, ts)
				if err != nil {
					select {
					case errCh <- err: // Record error once
					default:
					}
				} else {
					res.MetricName = metric // Add metric name
					mtx.Lock()
					metrics = append(metrics, res)
					mtx.Unlock()
				}
				wg.Done()
			}(metric, exp)
		}
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return metrics, nil
	}
}

func (p *prometheus) GetNamedMetricsOverTime(start, end time.Time, step time.Duration, o monitoring.QueryOption) ([]monitoring.Metric, error) {
	metrics := make([]monitoring.Metric, 0)
	var mtx sync.Mutex // guard metrics
	var wg sync.WaitGroup

	opts := monitoring.NewQueryOptions()
	o.Apply(opts)

	errCh := make(chan error)
	for _, metric := range opts.NamedMetrics {
		matched, _ := regexp.MatchString(opts.MetricFilter, metric)
		if matched {
			exp := makeExpression(metric, *opts)
			wg.Add(1)
			go func(metric, exp string) {
				res, err := p.rangeQuery(exp, start, end, step)
				if err != nil {
					select {
					case errCh <- err: // Record error once
					default:
					}
				} else {
					res.MetricName = metric // Add metric name
					mtx.Lock()
					metrics = append(metrics, res)
					mtx.Unlock()
				}
				wg.Done()
			}(metric, exp)
		}
	}

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		return metrics, nil
	}
}

func (p prometheus) query(exp string, ts time.Time) (monitoring.Metric, error) {
	params := &url.Values{}
	params.Set("time", ts.Format(time.RFC3339))
	params.Set("query", exp)

	u := fmt.Sprintf("%s/api/v1/query?%s", p.options.Endpoint, params.Encode())

	var m monitoring.Metric
	response, err := p.client.Get(u)
	if err != nil {
		return monitoring.Metric{}, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return monitoring.Metric{}, err
	}
	defer response.Body.Close()

	err = json.Unmarshal(body, m)
	if err != nil {
		return monitoring.Metric{}, err
	}

	return m, nil
}

func (p prometheus) rangeQuery(exp string, start, end time.Time, step time.Duration) (monitoring.Metric, error) {
	params := &url.Values{}
	params.Set("start", start.Format(time.RFC3339))
	params.Set("end", end.Format(time.RFC3339))
	params.Set("step", step.String())
	params.Set("query", exp)

	u := fmt.Sprintf("%s/api/v1/query?%s", p.options.Endpoint, params.Encode())

	var m monitoring.Metric
	response, err := p.client.Get(u)
	if err != nil {
		return monitoring.Metric{}, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return monitoring.Metric{}, err
	}
	defer response.Body.Close()

	err = json.Unmarshal(body, m)
	if err != nil {
		return monitoring.Metric{}, err
	}

	return m, nil
}

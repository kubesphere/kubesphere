package monitoring

import "time"

type Interface interface {
	GetMetrics(exprs []string, time time.Time) []Metric
	GetMetricsOverTime(exprs []string, start, end time.Time, step time.Duration) []Metric
	GetNamedMetrics(metrics []string, time time.Time, opt QueryOption) []Metric
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt QueryOption) []Metric
}

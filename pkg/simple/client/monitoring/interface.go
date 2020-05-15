package monitoring

import "time"

type Interface interface {
	GetMetric(expr string, time time.Time) Metric
	GetMetricOverTime(expr string, start, end time.Time, step time.Duration) Metric
	GetNamedMetrics(metrics []string, time time.Time, opt QueryOption) []Metric
	GetNamedMetricsOverTime(metrics []string, start, end time.Time, step time.Duration, opt QueryOption) []Metric
	GetMetadata(namespace string) []Metadata
	GetMetricLabels(expr string, start, end time.Time) MetricLabels
}

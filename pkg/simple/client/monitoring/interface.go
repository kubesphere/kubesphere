package monitoring

import "time"

const (
	StatusSuccess    = "success"
	StatusError      = "error"
	MetricTypeMatrix = "matrix"
	MetricTypeVector = "vector"
)

type Metric struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum"`
	Status     string `json:"status" description:"result status, one of error, success"`
	MetricData `json:"data" description:"actual metric result"`
	ErrorType  string `json:"errorType,omitempty"`
	Error      string `json:"error,omitempty"`
}

type MetricData struct {
	MetricType   string        `json:"resultType" description:"result type, one of matrix, vector"`
	MetricValues []MetricValue `json:"result" description:"metric data including labels, time series and values"`
}

type Point [2]float64

type MetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"time series labels"`
	Sample   Point             `json:"value,omitempty" description:"time series, values of vector type"`
	Series   []Point           `json:"values,omitempty" description:"time series, values of matrix type"`
}

type Interface interface {
	// The `stmts` defines statements, expressions or rules (eg. promql in Prometheus) for querying specific metrics.
	GetMetrics(stmts []string, time time.Time) ([]Metric, error)
	GetMetricsOverTime(stmts []string, start, end time.Time, step time.Duration) ([]Metric, error)

	// Get named metrics (eg. node_cpu_usage)
	GetNamedMetrics(time time.Time, opt QueryOption) ([]Metric, error)
	GetNamedMetricsOverTime(start, end time.Time, step time.Duration, opt QueryOption) ([]Metric, error)
}

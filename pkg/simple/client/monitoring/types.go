package monitoring

const (
	MetricTypeMatrix = "matrix"
	MetricTypeVector = "vector"
)

type Metric struct {
	MetricName string `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum"`
	MetricData `json:"data,omitempty" description:"actual metric result"`
	Error      string `json:"error,omitempty"`
}

type MetricData struct {
	MetricType   string        `json:"resultType,omitempty" description:"result type, one of matrix, vector"`
	MetricValues []MetricValue `json:"result,omitempty" description:"metric data including labels, time series and values"`
}

type Point [2]float64

type MetricValue struct {
	Metadata map[string]string `json:"metric,omitempty" description:"time series labels"`
	Sample   Point             `json:"value,omitempty" description:"time series, values of vector type"`
	Series   []Point           `json:"values,omitempty" description:"time series, values of matrix type"`
}

func (p Point) Timestamp() float64 {
	return p[0]
}

func (p Point) Value() float64 {
	return p[1]
}

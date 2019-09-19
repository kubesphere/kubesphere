package v1alpha2

// Prometheus query api response
type APIResponse struct {
	Status    string      `json:"status" description:"result status, one of error, success"`
	Data      QueryResult `json:"data" description:"actual metric result"`
	ErrorType string      `json:"errorType,omitempty"`
	Error     string      `json:"error,omitempty"`
	Warnings  []string    `json:"warnings,omitempty"`
}

// QueryResult includes result data from a query.
type QueryResult struct {
	ResultType string       `json:"resultType" description:"result type, one of matrix, vector"`
	Result     []QueryValue `json:"result" description:"metric data including labels, time series and values"`
}

// Time Series
type QueryValue struct {
	Metric map[string]string `json:"metric,omitempty" description:"time series labels"`
	Value  []interface{}     `json:"value,omitempty" description:"time series, values of vector type"`
	Values [][]interface{}   `json:"values,omitempty" description:"time series, values of matrix type"`
}

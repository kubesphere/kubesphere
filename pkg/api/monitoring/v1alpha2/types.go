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

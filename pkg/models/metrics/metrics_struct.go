/*
Copyright 2018 The KubeSphere Authors.

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

package metrics

type FormatedLevelMetric struct {
	MetricsLevel string           `json:"metrics_level"`
	Results      []FormatedMetric `json:"results"`
}

type FormatedMetric struct {
	MetricName string             `json:"metric_name, omitempty"`
	Status     string             `json:"status"`
	Data       FormatedMetricData `json:"data, omitempty"`
}

type FormatedMetricData struct {
	Result     []map[string]interface{} `json:"result"`
	ResultType string                   `json:"resultType"`
}

type CommonMetricsResult struct {
	Status string            `json:"status"`
	Data   CommonMetricsData `json:"data"`
}

type CommonMetricsData struct {
	Result     []CommonResultItem `json:"result"`
	ResultType string             `json:"resultType"`
}

type CommonResultItem struct {
	KubePodMetric KubePodMetric `json:"metric"`
	Value         interface{}   `json:"value"`
}

/**
"__name__": "kube_pod_info",
"created_by_kind": "\\u003cnone\\u003e",
"created_by_name": "\\u003cnone\\u003e",
"endpoint": "https-main",
"host_ip": "192.168.0.13",
"instance": "10.244.114.187:8443",
"job": "kube-state-metrics",
"namespace": "kube-system",
"node": "i-39p7faw6",
"pod": "cloud-controller-manager-i-39p7faw6",
"pod_ip": "192.168.0.13",
"service": "kube-state-metrics"
*/
type KubePodMetric struct {
	CreatedByKind string `json:"created_by_kind"`
	CreatedByName string `json:"created_by_name"`
	Namespace     string `json:"namespace"`
	Pod           string `json:"pod"`
}

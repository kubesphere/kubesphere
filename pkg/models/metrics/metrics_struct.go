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

/**
"metrics_level": "node",
"name": "i-9waiax0b",
"results": [{
		"metrics_name": "node_cpu_utilization",
		"status": "success",
		"data": {
			"resultType": "vector",
			"result": [{
				"metric": {
					"__name__": "node:node_cpu_utilisation:avg1m",
					"node": "i-9waiax0b"
				},
				"value": [1534834542.74, "0.04733333333327516"]
			}]
		}
	},
*/

type CommonMultipleMertics struct {
	MetricsLevel string `json:"metrics_level"`
	//Results interface{} `json:"results"`
	Results []CommonOneMetric `json:"results"`
}

type CommonOneMetric struct {
	MetricName string           `json:"metric_name"`
	Status     string           `json:"status"`
	Data       CommonMetricData `json:"data"`
}

type CommonMetricData struct {
	Result     []PodResultItem `json:"result"`
	ResultType string          `json:"resultType"`
}

type PodResultItem struct {
	Metric Metric      `json:"metric"`
	Value  interface{} `json:"value, omitempty"`
	Values interface{} `json:"values, omitempty"`
}

type Metric struct {
	Name      string `json:"__name__, omitempty"`
	Node      string `json:"node, omitempty"`
	Namespace string `json:"namespace, omitempty"`
	PodName   string `json:"pod_name, omitempty"`
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
	Name          string `json:"__name__"`
	CreatedByKind string `json:"created_by_kind"`
	CreatedByName string `json:"created_by_name"`
	EndPoint      string `json:"endpoint"`
	HostIP        string `json:"host_ip"`
	Instance      string `json:"instance"`
	Job           string `json:"job"`
	Namespace     string `json:"namespace"`
	Node          string `json:"node"`
	Pod           string `json:"pod"`
	PodIP         string `json:"pod_ip"`
	Service       string `json:"service"`
}

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

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
)

func getPodNameRegexInWorkload(request *restful.Request) string {
	promql := MakeWorkloadRule(request)
	res := client.SendPrometheusRequest(request, promql)
	data := []byte(res)
	var dat CommonMetricsResult
	jsonErr := json.Unmarshal(data, &dat)
	if jsonErr != nil {
		glog.Errorln("json parse failed", jsonErr)
	}
	var podNames []string
	for _, x := range dat.Data.Result {
		podName := x.KubePodMetric.Pod
		podNames = append(podNames, podName)
	}
	podNamesFilter := "^(" + strings.Join(podNames, "|") + ")$"
	return podNamesFilter
}

func MonitorWorkloadSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	podNamesFilter := getPodNameRegexInWorkload(request)
	newPromql := MakePodPromQL(request, []string{metricsName, nsName, "", "", podNamesFilter})
	podMetrics := client.SendPrometheusRequest(request, newPromql)
	cleanedJson := ReformatJson(podMetrics, metricsName)
	return cleanedJson
}

func MonitorPodSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	nsName := strings.Trim(request.PathParameter("ns_name"), " ")
	nodeID := strings.Trim(request.PathParameter("node_id"), " ")
	podName := strings.Trim(request.PathParameter("pod_name"), " ")
	podFilter := strings.Trim(request.QueryParameter("pods_filter"), " ")
	params := []string{metricsName, nsName, nodeID, podName, podFilter}
	promql := MakePodPromQL(request, params)
	if promql != "" {
		res := client.SendPrometheusRequest(request, promql)
		cleanedJson := ReformatJson(res, metricsName)
		return cleanedJson
	}
	return nil
}

func MonitorNamespaceSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	recordingRule := MakeNamespacePromQL(request, metricsName)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := ReformatJson(res, metricsName)
	return cleanedJson
}

func ReformatJson(metric string, metricsName string) *FormatedMetric {
	var formatMetric FormatedMetric
	err := json.Unmarshal([]byte(metric), &formatMetric)
	if err != nil {
		glog.Errorln("Unmarshal metric json failed", err)
	}
	if formatMetric.MetricName == "" {
		formatMetric.MetricName = metricsName
	}
	// retrive metrics success
	if formatMetric.Status == "success" {
		result := formatMetric.Data.Result
		for _, res := range result {
			metric, ok := res["metric"]
			me := metric.(map[string]interface{})
			if ok {
				delete(me, "__name__")
			}
		}
	}
	return &formatMetric
}

func collectNodeorClusterMetrics(request *restful.Request, metricsName string, ch chan<- *FormatedMetric) {
	metric := MonitorNodeorClusterSingleMetric(request, metricsName)
	ch <- metric
}

func collectNamespaceMetrics(request *restful.Request, metricsName string, ch chan<- *FormatedMetric) {
	metric := MonitorNamespaceSingleMetric(request, metricsName)
	ch <- metric
}

func collectWorkloadMetrics(request *restful.Request, metricsName string, ch chan<- *FormatedMetric) {
	metricsName = strings.TrimLeft(metricsName, "workload_")
	metric := MonitorWorkloadSingleMetric(request, metricsName)
	ch <- metric
}

func collectPodMetrics(request *restful.Request, metricsName string, ch chan<- *FormatedMetric) {
	metric := MonitorPodSingleMetric(request, metricsName)
	ch <- metric
}

func MonitorAllMetrics(request *restful.Request) FormatedLevelMetric {
	metricsName := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	if metricsName == "" {
		metricsName = ".*"
	}
	path := request.SelectedRoutePath()
	sourceType := path[strings.LastIndex(path, "/")+1 : len(path)-1]
	if strings.Contains(path, "workload") {
		sourceType = "workload"
	}
	var ch = make(chan *FormatedMetric, 10)
	for _, k := range MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed", err)
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			if sourceType == "node" || sourceType == "cluster" {
				go collectNodeorClusterMetrics(request, k, ch)
			} else if sourceType == "namespace" {
				go collectNamespaceMetrics(request, k, ch)
			} else if sourceType == "pod" {
				go collectPodMetrics(request, k, ch)
			} else if sourceType == "workload" {
				go collectWorkloadMetrics(request, k, ch)
			}
		}
	}
	var metricsArray []FormatedMetric
	var tempJson *FormatedMetric
	for _, k := range MetricsNames {
		bol, err := regexp.MatchString(metricsName, k)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed")
			continue
		}
		if strings.HasPrefix(k, sourceType) {
			tempJson = <-ch
			if tempJson != nil {
				metricsArray = append(metricsArray, *tempJson)
			}
		}
	}
	return FormatedLevelMetric{
		MetricsLevel: sourceType,
		Results:      metricsArray,
	}
}

func MonitorNodeorClusterSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	recordingRule := MakeNodeorClusterRule(request, metricsName)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := ReformatJson(res, metricsName)
	return cleanedJson
}

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

	"time"

	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/models"
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

// maybe this function is time consuming
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
	if formatMetric.Status == MetricStatusSuccess {
		result := formatMetric.Data.Result
		for _, res := range result {
			metric, ok := res[ResultItemMetric]
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

func collectWorkspaceMetrics(request *restful.Request, metricsName string, namespaceList []string, ch chan<- *FormatedMetric) {
	mertic := monitorWorkspaceSingleMertic(request, metricsName, namespaceList)
	ch <- mertic
}

func collectPodMetrics(request *restful.Request, metricsName string, ch chan<- *FormatedMetric) {
	metric := MonitorPodSingleMetric(request, metricsName)
	ch <- metric
}

func monitorWorkspaceSingleMertic(request *restful.Request, metricsName string, namespaceList []string) *FormatedMetric {
	namespaceRe2 := "^(" + strings.Join(namespaceList, "|") + ")$"
	newpromql := MakeWorkspacePromQL(metricsName, namespaceRe2)
	podMetrics := client.SendPrometheusRequest(request, newpromql)
	cleanedJson := ReformatJson(podMetrics, metricsName)
	return cleanedJson
}

func filterNamespace(request *restful.Request, namespaceList []string) []string {
	var newNSlist []string
	nsFilter := strings.Trim(request.QueryParameter("namespaces_filter"), " ")
	if nsFilter == "" {
		nsFilter = ".*"
	}
	for _, ns := range namespaceList {
		bol, _ := regexp.MatchString(nsFilter, ns)
		if bol {
			newNSlist = append(newNSlist, ns)
		}
	}
	return newNSlist
}

func MonitorAllMetrics(request *restful.Request) FormatedLevelMetric {
	metricsName := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	if metricsName == "" {
		metricsName = ".*"
	}
	path := request.SelectedRoutePath()
	sourceType := path[strings.LastIndex(path, "/")+1 : len(path)-1]
	if strings.Contains(path, MetricLevelWorkload) {
		sourceType = MetricLevelWorkload
	} else if strings.Contains(path, MetricLevelWorkspace) {
		sourceType = MetricLevelWorkspace
	}
	var ch = make(chan *FormatedMetric, 10)
	for _, metricName := range MetricsNames {
		bol, err := regexp.MatchString(metricsName, metricName)
		if !bol {
			continue
		}
		if err != nil {
			glog.Errorln("regex match failed", err)
			continue
		}
		if strings.HasPrefix(metricName, sourceType) {
			if sourceType == MetricLevelCluster || sourceType == MetricLevelNode {
				go collectNodeorClusterMetrics(request, metricName, ch)
			} else if sourceType == MetricLevelNamespace {
				go collectNamespaceMetrics(request, metricName, ch)
			} else if sourceType == MetricLevelPod {
				go collectPodMetrics(request, metricName, ch)
			} else if sourceType == MetricLevelWorkload {
				go collectWorkloadMetrics(request, metricName, ch)
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

func getWorkspacePodsCountMetrics(request *restful.Request, namespaces []string) *FormatedMetric {
	metricName := MetricNameNamespacePodCount
	var recordingRule = RulePromQLTmplMap[metricName]
	nsFilter := "^(" + strings.Join(namespaces, "|") + ")$"
	recordingRule = strings.Replace(recordingRule, "$1", nsFilter, -1)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := ReformatJson(res, metricName)
	return cleanedJson
}

func getWorkspaceWorkloadCountMetrics(namespaces []string) FormatedMetric {
	var wlQuotaMetrics models.ResourceQuota
	wlQuotaMetrics.NameSpace = strings.Join(namespaces, "|")
	wlQuotaMetrics.Data.Used = make(v1.ResourceList, 1)
	wlQuotaMetrics.Data.Hard = make(v1.ResourceList, 1)
	for _, ns := range namespaces {
		quotaMetric, err := models.GetNamespaceQuota(ns)
		if err != nil {
			glog.Errorln(err)
			continue
		}
		// sum all resources used along namespaces
		quotaUsed := quotaMetric.Data.Used
		for resourceName, quantity := range quotaUsed {
			if _, ok := wlQuotaMetrics.Data.Used[resourceName]; ok {
				tmpQuantity := wlQuotaMetrics.Data.Used[v1.ResourceName(resourceName)]
				tmpQuantity.Add(quantity)
				wlQuotaMetrics.Data.Used[v1.ResourceName(resourceName)] = tmpQuantity
			} else {
				wlQuotaMetrics.Data.Used[v1.ResourceName(resourceName)] = quantity.DeepCopy()
			}
		}

		// sum all resources hard along namespaces
		quotaHard := quotaMetric.Data.Hard
		for resourceName, quantity := range quotaHard {
			if _, ok := wlQuotaMetrics.Data.Hard[resourceName]; ok {
				tmpQuantity := wlQuotaMetrics.Data.Hard[v1.ResourceName(resourceName)]
				tmpQuantity.Add(quantity)
				wlQuotaMetrics.Data.Hard[v1.ResourceName(resourceName)] = tmpQuantity
			} else {
				wlQuotaMetrics.Data.Hard[v1.ResourceName(resourceName)] = quantity.DeepCopy()
			}
		}
	}
	wlMetrics := convertQuota2MetricStruct(&wlQuotaMetrics)
	return wlMetrics
}

func getSpecificMetricItem(timestamp int64, metricName string, kind string, count int, err error) FormatedMetric {
	var nsMetrics FormatedMetric
	nsMetrics.MetricName = metricName
	nsMetrics.Data.ResultType = ResultTypeVector
	resultItem := make(map[string]interface{})
	tmp := make(map[string]string)
	tmp[ResultItemMetricResource] = kind
	if err == nil {
		nsMetrics.Status = MetricStatusSuccess
	} else {
		nsMetrics.Status = MetricStatusError
		resultItem["errorinfo"] = err.Error()
	}

	resultItem[ResultItemMetric] = tmp
	resultItem[ResultItemValue] = []interface{}{timestamp, count}
	nsMetrics.Data.Result = make([]map[string]interface{}, 1)
	nsMetrics.Data.Result[0] = resultItem
	return nsMetrics
}

func MonitorNodeorClusterSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	// support cluster node statistic, include healthy nodes and unhealthy nodes
	var res string
	var fMetric FormatedMetric
	timestamp := int64(time.Now().Unix())

	if metricsName == MetricNameClusterHealthyNodeCount {
		onlineNodes, _ := getNodeHealthyConditionMetric()
		fMetric = getSpecificMetricItem(timestamp, MetricNameClusterHealthyNodeCount, "node_count", len(onlineNodes), nil)
	} else if metricsName == MetricNameClusterUnhealthyNodeCount {
		_, offlineNodes := getNodeHealthyConditionMetric()
		fMetric = getSpecificMetricItem(timestamp, MetricNameClusterUnhealthyNodeCount, "node_count", len(offlineNodes), nil)
	} else if metricsName == MetricNameClusterNodeCount {
		onlineNodes, offlineNodes := getNodeHealthyConditionMetric()
		fMetric = getSpecificMetricItem(timestamp, MetricNameClusterNodeCount, "node_count", len(onlineNodes)+len(offlineNodes), nil)
	} else {
		recordingRule := MakeNodeorClusterRule(request, metricsName)
		res = client.SendPrometheusRequest(request, recordingRule)
		fMetric = *ReformatJson(res, metricsName)
	}
	return &fMetric
}

func getNodeHealthyConditionMetric() ([]string, []string) {
	nodeList, err := client.NewK8sClient().CoreV1().Nodes().List(metaV1.ListOptions{})
	if err != nil {
		glog.Errorln(err)
		return nil, nil
	}
	var onlineNodes []string
	var offlineNodes []string
	for _, node := range nodeList.Items {
		nodeName := node.Labels["kubernetes.io/hostname"]
		nodeRole := node.Labels["role"]
		bol := true
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "Unknown" {
				bol = false
				break
			}
		}
		if nodeRole != "log" {
			if bol {
				// reachable node
				onlineNodes = append(onlineNodes, nodeName)
			} else {
				// unreachable node
				offlineNodes = append(offlineNodes, nodeName)
			}
		}
	}
	return onlineNodes, offlineNodes
}

func getExistingNamespace(namespaces []string) ([]string, []string) {
	namespaceMap, err := getAllNamespace()
	var existedNs []string
	var noneExistedNs []string
	if err != nil {
		return namespaces, nil
	}
	for _, ns := range namespaces {
		if _, ok := namespaceMap[ns]; ok {
			existedNs = append(existedNs, ns)
		} else {
			noneExistedNs = append(noneExistedNs, ns)
		}
	}
	return existedNs, noneExistedNs
}

func getAllNamespace() (map[string]int, error) {
	k8sClient := client.NewK8sClient()
	nsList, err := k8sClient.CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		glog.Errorln(err)
		return nil, err
	}
	namespaceMap := make(map[string]int)
	for _, item := range nsList.Items {
		namespaceMap[item.Name] = 0
	}
	return namespaceMap, nil
}

func MonitorWorkloadCount(request *restful.Request) FormatedMetric {
	namespace := strings.Trim(request.PathParameter("ns_name"), " ")

	quotaMetric, err := models.GetNamespaceQuota(namespace)
	fMetric := convertQuota2MetricStruct(quotaMetric)

	// whether the namespace in request parameters exists?
	namespaceMap, e := getAllNamespace()
	_, ok := namespaceMap[namespace]
	if e != nil {
		ok = true
	}
	if !ok || err != nil {
		fMetric.Status = MetricStatusError
		fMetric.Data.ResultType = ""
		errInfo := make(map[string]interface{})
		if err != nil {
			errInfo["errormsg"] = err.Error()
		} else {
			errInfo["errormsg"] = "namespace " + namespace + " does not exist"
		}
		fMetric.Data.Result = []map[string]interface{}{errInfo}
	}

	return fMetric
}

func convertQuota2MetricStruct(quotaMetric *models.ResourceQuota) FormatedMetric {
	var fMetric FormatedMetric
	fMetric.MetricName = MetricNameWorkloadCount
	fMetric.Status = MetricStatusSuccess
	fMetric.Data.ResultType = ResultTypeVector
	timestamp := int64(time.Now().Unix())
	var resultItems []map[string]interface{}

	hardMap := make(map[string]string)
	for resourceName, v := range quotaMetric.Data.Hard {
		hardMap[resourceName.String()] = v.String()
	}

	for resourceName, v := range quotaMetric.Data.Used {
		resultItem := make(map[string]interface{})
		tmp := make(map[string]string)
		tmp[ResultItemMetricResource] = resourceName.String()
		resultItem[ResultItemMetric] = tmp
		resultItem[ResultItemValue] = []interface{}{timestamp, hardMap[resourceName.String()], v.String()}
		resultItems = append(resultItems, resultItem)
	}

	fMetric.Data.Result = resultItems
	return fMetric
}

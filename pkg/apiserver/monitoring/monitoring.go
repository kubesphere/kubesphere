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
package monitoring

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/metrics"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func MonitorAllPodsOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorPod(request, response)
}

func MonitorSpecificPodOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorPod(request, response)
}

func MonitorAllPodsOnSpecificNode(request *restful.Request, response *restful.Response) {
	MonitorPod(request, response)
}

func MonitorSpecificPodOnSpecificNode(request *restful.Request, response *restful.Response) {
	MonitorPod(request, response)
}

func MonitorPod(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)
	podName := requestParams.PodName
	if podName != "" {
		requestParams.ResourcesFilter = fmt.Sprintf("^%s$", requestParams.PodName)
	}

	rawMetrics := metrics.GetPodLevelMetrics(requestParams)
	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
	response.WriteAsJson(pagedMetrics)
}

func MonitorAllContainersOnSpecificNode(request *restful.Request, response *restful.Response) {
	MonitorContainer(request, response)
}

func MonitorAllContainersOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorContainer(request, response)
}

func MonitorSpecificContainerOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorContainer(request, response)
}

func MonitorContainer(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)
	rawMetrics := metrics.GetContainerLevelMetrics(requestParams)
	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

	response.WriteAsJson(pagedMetrics)
}

func MonitorSpecificWorkload(request *restful.Request, response *restful.Response) {
	MonitorWorkload(request, response)
}

func MonitorAllWorkloadsOfSpecificKind(request *restful.Request, response *restful.Response) {
	MonitorWorkload(request, response)
}

func MonitorAllWorkloadsOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorWorkload(request, response)
}

func MonitorWorkload(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)

	rawMetrics := metrics.GetWorkloadLevelMetrics(requestParams)

	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

	response.WriteAsJson(pagedMetrics)

}

func MonitorAllWorkspaces(request *restful.Request, response *restful.Response) {

	requestParams := ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "statistics" {
		// merge multiple metric: all-devops, all-roles, all-projects...this api is designed for admin
		res := metrics.GetAllWorkspacesStatistics()
		response.WriteAsJson(res)

	} else {
		rawMetrics := metrics.MonitorAllWorkspaces(requestParams)

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorSpecificWorkspace(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)

	tp := requestParams.Tp
	if tp == "rank" {
		// multiple
		rawMetrics := metrics.GetWorkspaceLevelMetrics(requestParams)

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)

		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
		response.WriteAsJson(pagedMetrics)

	} else if tp == "statistics" {
		wsName := requestParams.WsName

		// merge multiple metric: devops, roles, projects...
		res := metrics.MonitorOneWorkspaceStatistics(wsName)
		response.WriteAsJson(res)
	} else {
		res := metrics.GetWorkspaceLevelMetrics(requestParams)
		response.WriteAsJson(res)
	}
}

func MonitorAllNamespaces(request *restful.Request, response *restful.Response) {
	MonitorNamespace(request, response)
}

func MonitorSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorNamespace(request, response)
}

func MonitorNamespace(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)
	// multiple
	rawMetrics := metrics.GetNamespaceLevelMetrics(requestParams)

	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
	response.WriteAsJson(pagedMetrics)
}

func MonitorCluster(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		prometheusClient, err := client.ClientSets().Prometheus()
		if err != nil {
			if _, ok := err.(client.ClientSetNotEnabledError); ok {
				klog.Error("monitoring is not enabled")
				return
			} else {
				klog.Errorf("get prometheus client failed %+v", err)
			}
		}

		// single
		queryType, params := metrics.AssembleClusterMetricRequestInfo(requestParams, metricName)
		metricsStr := prometheusClient.SendMonitoringRequest(queryType, params)
		res := metrics.ReformatJson(metricsStr, metricName, map[string]string{metrics.MetricLevelCluster: "local"})

		response.WriteAsJson(res)
	} else {
		// multiple
		res := metrics.GetClusterLevelMetrics(requestParams)
		response.WriteAsJson(res)
	}
}

func MonitorAllNodes(request *restful.Request, response *restful.Response) {
	MonitorNode(request, response)
}

func MonitorSpecificNode(request *restful.Request, response *restful.Response) {
	MonitorNode(request, response)
}

func MonitorNode(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)

	metricName := requestParams.MetricsName
	if metricName != "" {
		prometheusClient, err := client.ClientSets().Prometheus()
		if err != nil {
			if _, ok := err.(client.ClientSetNotEnabledError); ok {
				klog.Error("monitoring is not enabled")
				return
			} else {
				klog.Errorf("get prometheus client failed %+v", err)
			}
		}
		// single
		queryType, params := metrics.AssembleNodeMetricRequestInfo(requestParams, metricName)
		metricsStr := prometheusClient.SendMonitoringRequest(queryType, params)
		res := metrics.ReformatJson(metricsStr, metricName, map[string]string{metrics.MetricLevelNode: ""})
		// The raw node-exporter result doesn't include ip address information
		// Thereby, append node ip address to .data.result[].metric

		nodeAddress := metrics.GetNodeAddressInfo()
		metrics.AddNodeAddressMetric(res, nodeAddress)

		response.WriteAsJson(res)
	} else {
		// multiple
		rawMetrics := metrics.GetNodeLevelMetrics(requestParams)
		nodeAddress := metrics.GetNodeAddressInfo()

		for i := 0; i < len(rawMetrics.Results); i++ {
			metrics.AddNodeAddressMetric(&rawMetrics.Results[i], nodeAddress)
		}

		// sorting
		sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
		// paging
		pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)

		response.WriteAsJson(pagedMetrics)
	}
}

func MonitorAllPVCsOfSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorPVC(request, response)
}

func MonitorAllPVCsOfSpecificStorageClass(request *restful.Request, response *restful.Response) {
	MonitorPVC(request, response)
}

func MonitorSpecificPVCofSpecificNamespace(request *restful.Request, response *restful.Response) {
	MonitorPVC(request, response)
}

func MonitorPVC(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)
	pvcName := requestParams.PVCName
	if pvcName != "" {
		requestParams.ResourcesFilter = fmt.Sprintf("^%s$", requestParams.PVCName)
	}

	rawMetrics := metrics.GetPVCLevelMetrics(requestParams)
	// sorting
	sortedMetrics, maxMetricCount := metrics.Sort(requestParams.SortMetricName, requestParams.SortType, rawMetrics)
	// paging
	pagedMetrics := metrics.Page(requestParams.PageNum, requestParams.LimitNum, sortedMetrics, maxMetricCount)
	response.WriteAsJson(pagedMetrics)
}

func MonitorComponent(request *restful.Request, response *restful.Response) {
	requestParams := ParseMonitoringRequestParams(request)

	if requestParams.MetricsFilter == "" {
		requestParams.MetricsFilter = requestParams.ComponentName + "_.*"
	}

	rawMetrics := metrics.GetComponentLevelMetrics(requestParams)

	response.WriteAsJson(rawMetrics)
}

func ParseMonitoringRequestParams(request *restful.Request) *metrics.MonitoringRequestParams {
	instantTime := strings.Trim(request.QueryParameter("time"), " ")
	start := strings.Trim(request.QueryParameter("start"), " ")
	end := strings.Trim(request.QueryParameter("end"), " ")
	step := strings.Trim(request.QueryParameter("step"), " ")
	timeout := strings.Trim(request.QueryParameter("timeout"), " ")

	sortMetricName := strings.Trim(request.QueryParameter("sort_metric"), " ")
	sortType := strings.Trim(request.QueryParameter("sort_type"), " ")
	pageNum := strings.Trim(request.QueryParameter("page"), " ")
	limitNum := strings.Trim(request.QueryParameter("limit"), " ")
	tp := strings.Trim(request.QueryParameter("type"), " ")

	metricsFilter := strings.Trim(request.QueryParameter("metrics_filter"), " ")
	resourcesFilter := strings.Trim(request.QueryParameter("resources_filter"), " ")

	metricsName := strings.Trim(request.QueryParameter("metrics_name"), " ")
	workloadName := strings.Trim(request.PathParameter("workload"), " ")

	nodeId := strings.Trim(request.PathParameter("node"), " ")
	wsName := strings.Trim(request.PathParameter("workspace"), " ")
	nsName := strings.Trim(request.PathParameter("namespace"), " ")
	podName := strings.Trim(request.PathParameter("pod"), " ")
	pvcName := strings.Trim(request.PathParameter("pvc"), " ")
	storageClassName := strings.Trim(request.PathParameter("storageclass"), " ")
	containerName := strings.Trim(request.PathParameter("container"), " ")
	workloadKind := strings.Trim(request.PathParameter("kind"), " ")
	componentName := strings.Trim(request.PathParameter("component"), " ")

	var requestParams = metrics.MonitoringRequestParams{
		SortMetricName:   sortMetricName,
		SortType:         sortType,
		PageNum:          pageNum,
		LimitNum:         limitNum,
		Tp:               tp,
		MetricsFilter:    metricsFilter,
		ResourcesFilter:  resourcesFilter,
		MetricsName:      metricsName,
		WorkloadName:     workloadName,
		NodeId:           nodeId,
		WsName:           wsName,
		NsName:           nsName,
		PodName:          podName,
		PVCName:          pvcName,
		StorageClassName: storageClassName,
		ContainerName:    containerName,
		WorkloadKind:     workloadKind,
		ComponentName:    componentName,
	}

	if timeout == "" {
		timeout = metrics.DefaultQueryTimeout
	}
	if step == "" {
		step = metrics.DefaultQueryStep
	}
	// Whether query or query_range request
	u := url.Values{}

	if start != "" && end != "" {

		u.Set("start", convertTimeGranularity(start))
		u.Set("end", convertTimeGranularity(end))
		u.Set("step", step)
		u.Set("timeout", timeout)

		// range query start time must be greater than the namespace creation time
		if nsName != "" {
			nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
			ns, err := nsLister.Get(nsName)
			if err == nil {
				queryStartTime := u.Get("start")
				nsCreationTime := strconv.FormatInt(ns.CreationTimestamp.Unix(), 10)
				if nsCreationTime > queryStartTime {
					u.Set("start", nsCreationTime)
				}
			}
		}

		requestParams.QueryType = metrics.RangeQueryType
		requestParams.Params = u

		return &requestParams
	}
	if instantTime != "" {
		u.Set("time", instantTime)
		u.Set("timeout", timeout)
		requestParams.QueryType = metrics.DefaultQueryType
		requestParams.Params = u
		return &requestParams
	} else {
		u.Set("timeout", timeout)
		requestParams.QueryType = metrics.DefaultQueryType
		requestParams.Params = u
		return &requestParams
	}
}

func convertTimeGranularity(ts string) string {
	timeFloat, err := strconv.ParseFloat(ts, 64)
	if err != nil {
		klog.Errorf("convert second timestamp %s to minute timestamp failed", ts)
		return strconv.FormatInt(int64(time.Now().Unix()), 10)
	}
	timeInt := int64(timeFloat)
	// convert second timestamp to minute timestamp
	secondTime := time.Unix(timeInt, 0).Truncate(time.Minute).Unix()
	return strconv.FormatInt(secondTime, 10)
}

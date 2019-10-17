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

package metrics

import (
	"fmt"
	"github.com/json-iterator/go"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"kubesphere.io/kubesphere/pkg/models/workspaces"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

func GetClusterMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range clusterMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForCluster(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())
				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelCluster,
		Results:      apiResponse,
	}
}

func GetNodeMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range nodeMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForNode(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name, node_ip, node_role to each metric result item
				// resouce_name serves as a unique identifier for the monitored resource
				// it will be used during metrics sorting
				for _, item := range response.Data.Result {
					nodeName := item.Metric["node"]
					item.Metric["resource_name"] = nodeName
					item.Metric["node_ip"], item.Metric["node_role"] = getNodeAddressAndRole(nodeName)
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelNode,
		Results:      apiResponse,
	}
}

func GetWorkspaceMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range workspaceMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForWorkspace(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["label_kubesphere_io_workspace"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelWorkspace,
		Results:      apiResponse,
	}
}

func GetNamespaceMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range namespaceMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForNamespace(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["namespace"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelNamespace,
		Results:      apiResponse,
	}
}

func GetWorkloadMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range workloadMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForWorkload(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["workload"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelWorkload,
		Results:      apiResponse,
	}
}

func GetPodMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range podMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForPod(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["pod_name"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelPod,
		Results:      apiResponse,
	}
}

func GetContainerMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range containerMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForContainer(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["container_name"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelContainer,
		Results:      apiResponse,
	}
}

func GetPVCMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range pvcMetrics {
		matched, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matched {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForPVC(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SPrometheus(params.QueryType, v.Encode())

				// add label resouce_name
				for _, item := range response.Data.Result {
					item.Metric["resource_name"] = item.Metric["persistentvolumeclaim"]
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelPVC,
		Results:      apiResponse,
	}
}

func GetComponentMetrics(params RequestParams) *Response {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	ch := make(chan APIResponse, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// for each metric, make PromQL expression and send the request to Prometheus servers
	for _, metricName := range componentMetrics {
		matchComponentName, _ := regexp.MatchString(params.ComponentName, metricName)
		matchMetricsFilter, _ := regexp.MatchString(params.MetricsFilter, metricName)
		if matchComponentName && matchMetricsFilter {
			wg.Add(1)
			go func(metricName string, params RequestParams) {
				exp := makePromqlForComponent(metricName, params)
				v := url.Values{}
				for key, value := range params.QueryParams {
					v[key] = value
				}
				v.Set("query", exp)
				response := client.QueryToK8SSystemPrometheus(params.QueryType, v.Encode())

				// add node address when queried metric is etcd_server_list
				if metricName == "etcd_server_list" {
					for _, item := range response.Data.Result {
						item.Metric["node_name"] = getNodeName(item.Metric["node_ip"])
					}
				}

				ch <- APIResponse{
					MetricName:  metricName,
					APIResponse: response,
				}
				wg.Done()
			}(metricName, params)
		}
	}
	wg.Wait()
	close(ch)

	var apiResponse []APIResponse
	for e := range ch {
		apiResponse = append(apiResponse, e)
	}

	return &Response{
		MetricsLevel: MonitorLevelComponent,
		Results:      apiResponse,
	}
}

func makePromqlForCluster(metricName string, _ RequestParams) string {
	return metricsPromqlMap[metricName]
}

func makePromqlForNode(metricName string, params RequestParams) string {
	var rule = metricsPromqlMap[metricName]
	var nodeSelector string

	if params.NodeName != "" {
		nodeSelector = fmt.Sprintf(`node="%s"`, params.NodeName)
	} else {
		nodeSelector = fmt.Sprintf(`node=~"%s"`, params.ResourcesFilter)
	}

	return strings.Replace(rule, "$1", nodeSelector, -1)
}

func makePromqlForWorkspace(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var workspaceSelector string

	if params.WorkspaceName != "" {
		workspaceSelector = fmt.Sprintf(`label_kubesphere_io_workspace="%s"`, params.WorkspaceName)
	} else {
		workspaceSelector = fmt.Sprintf(`label_kubesphere_io_workspace=~"%s", label_kubesphere_io_workspace!=""`, params.ResourcesFilter)
	}

	return strings.Replace(exp, "$1", workspaceSelector, -1)
}

func makePromqlForNamespace(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var namespaceSelector string

	// For monitoring namespaces in the specific workspace
	// GET /workspaces/{workspace}/namespaces
	if params.WorkspaceName != "" {
		namespaceSelector = fmt.Sprintf(`label_kubesphere_io_workspace="%s", namespace=~"%s"`, params.WorkspaceName, params.ResourcesFilter)
		return strings.Replace(exp, "$1", namespaceSelector, -1)
	}

	// For monitoring the specific namespaces
	// GET /namespaces/{namespace} or
	// GET /namespaces
	if params.NamespaceName != "" {
		namespaceSelector = fmt.Sprintf(`namespace="%s"`, params.NamespaceName)
	} else {
		namespaceSelector = fmt.Sprintf(`namespace=~"%s"`, params.ResourcesFilter)
	}
	return strings.Replace(exp, "$1", namespaceSelector, -1)
}

func makePromqlForWorkload(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var kind, kindSelector, workloadSelector string

	switch params.WorkloadKind {
	case "deployment":
		kind = Deployment
		kindSelector = fmt.Sprintf(`namespace="%s", deployment!="", deployment=~"%s"`, params.NamespaceName, params.ResourcesFilter)
	case "statefulset":
		kind = StatefulSet
		kindSelector = fmt.Sprintf(`namespace="%s", statefulset!="", statefulset=~"%s"`, params.NamespaceName, params.ResourcesFilter)
	case "daemonset":
		kind = DaemonSet
		kindSelector = fmt.Sprintf(`namespace="%s", daemonset!="", daemonset=~"%s"`, params.NamespaceName, params.ResourcesFilter)
	default:
		kind = ".*"
		kindSelector = fmt.Sprintf(`namespace="%s"`, params.NamespaceName)
	}

	workloadSelector = fmt.Sprintf(`namespace="%s", workload=~"%s:%s"`, params.NamespaceName, kind, params.ResourcesFilter)
	return strings.NewReplacer("$1", workloadSelector, "$2", kindSelector).Replace(exp)
}

func makePromqlForPod(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var podSelector, workloadSelector string

	// For monitoriong pods of the specific workload
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods
	if params.WorkloadName != "" {
		switch params.WorkloadKind {
		case "deployment":
			workloadSelector = fmt.Sprintf(`owner_kind="ReplicaSet", owner_name=~"^%s-[^-]{1,10}$"`, params.WorkloadName)
		case "statefulset":
			workloadSelector = fmt.Sprintf(`owner_kind="StatefulSet", owner_name="%s"`, params.WorkloadName)
		case "daemonset":
			workloadSelector = fmt.Sprintf(`owner_kind="DaemonSet", owner_name="%s"`, params.WorkloadName)
		}
	}

	// For monitoring pods in the specific namespace
	// GET /namespaces/{namespace}/workloads/{kind}/{workload}/pods or
	// GET /namespaces/{namespace}/pods/{pod} or
	// GET /namespaces/{namespace}/pods
	if params.NamespaceName != "" {
		if params.PodName != "" {
			podSelector = fmt.Sprintf(`pod="%s", namespace="%s"`, params.PodName, params.NamespaceName)
		} else {
			podSelector = fmt.Sprintf(`pod=~"%s", namespace="%s"`, params.ResourcesFilter, params.NamespaceName)
		}
	}

	// For monitoring pods on the specific node
	// GET /nodes/{node}/pods/{pod}
	if params.NodeName != "" {
		if params.PodName != "" {
			podSelector = fmt.Sprintf(`pod="%s", node="%s"`, params.PodName, params.NodeName)
		} else {
			podSelector = fmt.Sprintf(`pod=~"%s", node="%s"`, params.ResourcesFilter, params.NodeName)
		}
	}

	return strings.NewReplacer("$1", workloadSelector, "$2", podSelector).Replace(exp)
}

func makePromqlForContainer(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var containerSelector string

	if params.ContainerName != "" {
		containerSelector = fmt.Sprintf(`pod_name="%s", namespace="%s", container_name="%s"`, params.PodName, params.NamespaceName, params.ContainerName)
	} else {
		containerSelector = fmt.Sprintf(`pod_name="%s", namespace="%s", container_name=~"%s"`, params.PodName, params.NamespaceName, params.ResourcesFilter)
	}

	return strings.Replace(exp, "$1", containerSelector, -1)
}

func makePromqlForPVC(metricName string, params RequestParams) string {
	var exp = metricsPromqlMap[metricName]
	var pvcSelector string

	// For monitoring persistentvolumeclaims in the specific namespace
	// GET /namespaces/{namespace}/persistentvolumeclaims/{persistentvolumeclaim} or
	// GET /namespaces/{namespace}/persistentvolumeclaims
	if params.NamespaceName != "" {
		if params.PVCName != "" {
			pvcSelector = fmt.Sprintf(`namespace="%s", persistentvolumeclaim="%s"`, params.NamespaceName, params.PVCName)
		} else {
			pvcSelector = fmt.Sprintf(`namespace="%s", persistentvolumeclaim=~"%s"`, params.NamespaceName, params.ResourcesFilter)
		}
		return strings.Replace(exp, "$1", pvcSelector, -1)
	}

	// For monitoring persistentvolumeclaims of the specific storageclass
	// GET /storageclasses/{storageclass}/persistentvolumeclaims
	if params.StorageClassName != "" {
		pvcSelector = fmt.Sprintf(`storageclass="%s", persistentvolumeclaim=~"%s"`, params.StorageClassName, params.ResourcesFilter)
	}
	return strings.Replace(exp, "$1", pvcSelector, -1)
}

func makePromqlForComponent(metricName string, _ RequestParams) string {
	return metricsPromqlMap[metricName]
}

func GetClusterStatistics() *Response {

	now := time.Now().Unix()

	var metricsArray []APIResponse
	workspaceStats := APIResponse{MetricName: MetricClusterWorkspaceCount}
	devopsStats := APIResponse{MetricName: MetricClusterDevopsCount}
	namespaceStats := APIResponse{MetricName: MetricClusterNamespaceCount}
	accountStats := APIResponse{MetricName: MetricClusterAccountCount}

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		num, err := workspaces.WorkspaceCount()
		if err != nil {
			klog.Errorln(err)
			workspaceStats.Status = "error"
		} else {
			workspaceStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		num, err := workspaces.GetAllDevOpsProjectsNums()
		if err != nil {
			if _, notEnabled := err.(cs.ClientSetNotEnabledError); !notEnabled {
				klog.Errorln(err)
			}
			devopsStats.Status = "error"
		} else {
			devopsStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		num, err := workspaces.GetAllProjectNums()
		if err != nil {
			klog.Errorln(err)
			namespaceStats.Status = "error"
		} else {
			namespaceStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		ret, err := cs.ClientSets().KubeSphere().ListUsers()
		if err != nil {
			klog.Errorln(err)
			accountStats.Status = "error"
		} else {
			accountStats.withMetricResult(now, ret.TotalCount)
		}
		wg.Done()
	}()

	wg.Wait()

	metricsArray = append(metricsArray, workspaceStats, devopsStats, namespaceStats, accountStats)

	return &Response{
		MetricsLevel: MonitorLevelCluster,
		Results:      metricsArray,
	}
}

func GetWorkspaceStatistics(workspaceName string) *Response {

	now := time.Now().Unix()

	var metricsArray []APIResponse
	namespaceStats := APIResponse{MetricName: MetricWorkspaceNamespaceCount}
	devopsStats := APIResponse{MetricName: MetricWorkspaceDevopsCount}
	memberStats := APIResponse{MetricName: MetricWorkspaceMemberCount}
	roleStats := APIResponse{MetricName: MetricWorkspaceRoleCount}

	wg := sync.WaitGroup{}
	wg.Add(4)

	go func() {
		num, err := workspaces.WorkspaceNamespaceCount(workspaceName)
		if err != nil {
			klog.Errorln(err)
			namespaceStats.Status = "error"
		} else {
			namespaceStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		num, err := workspaces.GetDevOpsProjectsCount(workspaceName)
		if err != nil {
			if _, notEnabled := err.(cs.ClientSetNotEnabledError); !notEnabled {
				klog.Errorln(err)
			}
			devopsStats.Status = "error"
		} else {
			devopsStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		num, err := workspaces.WorkspaceUserCount(workspaceName)
		if err != nil {
			klog.Errorln(err)
			memberStats.Status = "error"
		} else {
			memberStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	go func() {
		num, err := workspaces.GetOrgRolesCount(workspaceName)
		if err != nil {
			klog.Errorln(err)
			roleStats.Status = "error"
		} else {
			roleStats.withMetricResult(now, num)
		}
		wg.Done()
	}()

	wg.Wait()

	metricsArray = append(metricsArray, namespaceStats, devopsStats, memberStats, roleStats)

	return &Response{
		MetricsLevel: MonitorLevelWorkspace,
		Results:      metricsArray,
	}
}

func (response *APIResponse) withMetricResult(time int64, value int) {
	response.Status = "success"
	response.Data = v1alpha2.QueryResult{
		ResultType: "vector",
		Result: []v1alpha2.QueryValue{
			{
				Value: []interface{}{time, value},
			},
		},
	}
}

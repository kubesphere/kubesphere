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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/informers"
	"net/url"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/json-iterator/go"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/models/workspaces"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
)

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

const (
	ChannelMaxCapacityWorkspaceMetric = 800
	ChannelMaxCapacity                = 100
)

type FormatedLevelMetric struct {
	MetricsLevel string           `json:"metrics_level" description:"metric level, eg. cluster"`
	Results      []FormatedMetric `json:"results" description:"actual array of results"`
	CurrentPage  int              `json:"page,omitempty" description:"current page returned"`
	TotalPage    int              `json:"total_page,omitempty" description:"total number of pages"`
	TotalItem    int              `json:"total_item,omitempty" description:"page size"`
}

type FormatedMetric struct {
	MetricName string             `json:"metric_name,omitempty" description:"metric name, eg. scheduler_up_sum"`
	Status     string             `json:"status" description:"result status, one of error, success"`
	Data       FormatedMetricData `json:"data,omitempty" description:"actual metric result"`
}

type FormatedMetricData struct {
	Result     []map[string]interface{} `json:"result" description:"metric data including metric metadata, time points and values"`
	ResultType string                   `json:"resultType" description:"result type, one of matrix, vector"`
}

type MetricResultValues []MetricResultValue

type MetricResultValue struct {
	timestamp float64
	value     string
}

type MetricItem struct {
	MetricLabel map[string]string `json:"metric"`
	Value       []interface{}     `json:"value"`
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

type KubePodMetric struct {
	CreatedByKind string `json:"created_by_kind"`
	CreatedByName string `json:"created_by_name"`
	Namespace     string `json:"namespace"`
	Pod           string `json:"pod"`
}

type ComponentStatus struct {
	Name            string               `json:"metric_name,omitempty"`
	Namespace       string               `json:"namespace,omitempty"`
	Labels          map[string]string    `json:"labels,omitempty"`
	ComponentStatus []OneComponentStatus `json:"component"`
}

type OneComponentStatus struct {
	// Valid value: "Healthy"
	Type string `json:"type"`
	// Valid values for "Healthy": "True", "False", or "Unknown".
	Status string `json:"status"`
	// Message about the condition for a component.
	Message string `json:"message,omitempty"`
	// Condition error code for a component.
	Error string `json:"error,omitempty"`
}

func getAllWorkspaceNames(formatedMetric *FormatedMetric) map[string]int {

	var wsMap = make(map[string]int)

	for i := 0; i < len(formatedMetric.Data.Result); i++ {
		// metricDesc needs clear naming
		metricDesc := formatedMetric.Data.Result[i][ResultItemMetric]
		metricDescMap, ensure := metricDesc.(map[string]interface{})
		if ensure {
			if wsLabel, exist := metricDescMap[WorkspaceJoinedKey]; exist {
				wsMap[wsLabel.(string)] = 1
			}
		}
	}
	return wsMap
}

func getAllWorkspaces() map[string]int {

	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		return nil
	}

	paramValues := make(url.Values)
	paramValues.Set("query", WorkspaceNamespaceLabelRule)
	params := paramValues.Encode()
	res := client.SendSecondaryMonitoringRequest(DefaultQueryType, params)

	metric := ReformatJson(res, "", map[string]string{"workspace": "workspace"})

	return getAllWorkspaceNames(metric)
}

func getPodNameRegexInWorkload(res, filter string) string {

	data := []byte(res)
	var dat CommonMetricsResult
	jsonErr := jsonIter.Unmarshal(data, &dat)
	if jsonErr != nil {
		klog.Errorln("json parse failed", jsonErr.Error(), res)
	}
	var podNames []string

	for _, item := range dat.Data.Result {
		podName := item.KubePodMetric.Pod

		if filter != "" {
			if bol, _ := regexp.MatchString(filter, podName); bol {
				podNames = append(podNames, podName)
			}
		} else {
			podNames = append(podNames, podName)
		}

	}

	podNamesFilter := "^(" + strings.Join(podNames, "|") + ")$"
	return podNamesFilter
}

func unifyMetricHistoryTimeRange(fmtMetrics *FormatedMetric) {

	defer func() {
		if err := recover(); err != nil {
			klog.Errorln(err)
			debug.PrintStack()
		}
	}()

	var timestampMap = make(map[float64]bool)

	if fmtMetrics.Data.ResultType == ResultTypeMatrix {
		for i := range fmtMetrics.Data.Result {
			values, exist := fmtMetrics.Data.Result[i][ResultItemValues]
			if exist {
				valueArray, sure := values.([]interface{})
				if sure {
					for j := range valueArray {
						timeAndValue := valueArray[j].([]interface{})
						timestampMap[float64(timeAndValue[0].(uint64))] = true
					}
				}
			}
		}
	}

	timestampArray := make([]float64, len(timestampMap))
	i := 0
	for timestamp := range timestampMap {
		timestampArray[i] = timestamp
		i++
	}
	sort.Float64s(timestampArray)

	if fmtMetrics.Data.ResultType == ResultTypeMatrix {
		for i := 0; i < len(fmtMetrics.Data.Result); i++ {

			values, exist := fmtMetrics.Data.Result[i][ResultItemValues]
			if exist {
				valueArray, sure := values.([]interface{})
				if sure {

					formatValueArray := make([][]interface{}, len(timestampArray))
					j := 0

					for k := range timestampArray {
						valueItem, sure := valueArray[j].([]interface{})
						if sure && float64(valueItem[0].(uint64)) == timestampArray[k] {
							formatValueArray[k] = []interface{}{int64(timestampArray[k]), valueItem[1]}
							j++
						} else {
							formatValueArray[k] = []interface{}{int64(timestampArray[k]), "-1"}
						}
					}
					fmtMetrics.Data.Result[i][ResultItemValues] = formatValueArray
				}
			}
		}
	}
}

func AssembleSpecificWorkloadMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string, bool) {

	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return "", "", false
	}

	nsName := monitoringRequest.NsName
	wlName := monitoringRequest.WorkloadName
	podsFilter := monitoringRequest.ResourcesFilter

	rule := MakeSpecificWorkloadRule(monitoringRequest.WorkloadKind, wlName, nsName)
	paramValues := monitoringRequest.Params
	params := makeRequestParamString(rule, paramValues)

	res := client.SendMonitoringRequest(DefaultQueryType, params)

	podNamesFilter := getPodNameRegexInWorkload(res, podsFilter)

	queryType := monitoringRequest.QueryType
	rule = MakePodPromQL(metricName, nsName, "", "", podNamesFilter)
	params = makeRequestParamString(rule, paramValues)

	return queryType, params, rule == ""
}

func AssembleAllWorkloadMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params

	rule := MakeWorkloadPromQL(metricName, monitoringRequest.NsName, monitoringRequest.ResourcesFilter, monitoringRequest.WorkloadKind)

	params := makeRequestParamString(rule, paramValues)
	return queryType, params
}

func AssemblePodMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string, bool) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params

	rule := MakePodPromQL(metricName, monitoringRequest.NsName, monitoringRequest.NodeId, monitoringRequest.PodName, monitoringRequest.ResourcesFilter)
	params := makeRequestParamString(rule, paramValues)
	return queryType, params, rule == ""
}

func AssemblePVCMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string, bool) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params

	rule := MakePVCPromQL(metricName, monitoringRequest.NsName, monitoringRequest.PVCName, monitoringRequest.StorageClassName, monitoringRequest.ResourcesFilter)
	params := makeRequestParamString(rule, paramValues)
	return queryType, params, rule == ""
}

func GetNodeAddressInfo() *map[string][]v1.NodeAddress {
	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	nodes, err := nodeLister.List(labels.Everything())

	if err != nil {
		klog.Errorln(err.Error())
	}

	var nodeAddress = make(map[string][]v1.NodeAddress)

	for _, node := range nodes {
		nodeAddress[node.Name] = node.Status.Addresses
	}
	return &nodeAddress
}

func AddNodeAddressMetric(nodeMetric *FormatedMetric, nodeAddress *map[string][]v1.NodeAddress) {

	for i := 0; i < len(nodeMetric.Data.Result); i++ {
		metricDesc := nodeMetric.Data.Result[i][ResultItemMetric]
		metricDescMap, ensure := metricDesc.(map[string]interface{})
		if ensure {
			if nodeId, exist := metricDescMap[ResultItemMetricResourceName]; exist {
				addr, exist := (*nodeAddress)[nodeId.(string)]
				if exist {
					metricDescMap["address"] = addr
				}
			}
		}
	}
}

func MonitorContainer(monitoringRequest *MonitoringRequestParams, metricName string) *FormatedMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	queryType, params := AssembleContainerMetricRequestInfo(monitoringRequest, metricName)
	metricsStr := client.SendMonitoringRequest(queryType, params)
	res := ReformatJson(metricsStr, metricName, map[string]string{MetricLevelContainerName: ""})
	return res
}

func AssembleContainerMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params
	rule := MakeContainerPromQL(monitoringRequest.NsName, monitoringRequest.NodeId, monitoringRequest.PodName, monitoringRequest.ContainerName, metricName, monitoringRequest.ResourcesFilter)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleNamespaceMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params
	rule := MakeNamespacePromQL(monitoringRequest.NsName, monitoringRequest.ResourcesFilter, metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleNamespaceMetricRequestInfoByNamesapce(monitoringRequest *MonitoringRequestParams, namespace string, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params
	rule := MakeNamespacePromQL(namespace, monitoringRequest.ResourcesFilter, metricName)

	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest *MonitoringRequestParams, namespaceList []string, workspace string, metricName string) (string, string) {

	nsFilter := "^(" + strings.Join(namespaceList, "|") + ")$"

	queryType := monitoringRequest.QueryType

	rule := MakeSpecificWorkspacePromQL(metricName, nsFilter, workspace)
	paramValues := monitoringRequest.Params
	params := makeRequestParamString(rule, paramValues)
	return queryType, params
}

func AssembleAllWorkspaceMetricRequestInfo(monitoringRequest *MonitoringRequestParams, namespaceList []string, metricName string) (string, string) {
	var nsFilter = "^()$"

	if namespaceList != nil {
		nsFilter = "^(" + strings.Join(namespaceList, "|") + ")$"
	}

	queryType := monitoringRequest.QueryType

	rule := MakeAllWorkspacesPromQL(metricName, nsFilter)
	paramValues := monitoringRequest.Params
	params := makeRequestParamString(rule, paramValues)
	return queryType, params
}

func makeRequestParamString(rule string, paramValues url.Values) string {

	defer func() {
		if err := recover(); err != nil {
			klog.Errorln(err)
			debug.PrintStack()
		}
	}()

	var values = make(url.Values)
	for key, v := range paramValues {
		values.Set(key, v[0])
	}

	values.Set("query", rule)

	params := values.Encode()

	return params
}

func filterNamespace(nsFilter string, namespaceList []string) []string {
	var newNSlist []string
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

func MonitorAllWorkspaces(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	metricsFilter := monitoringRequest.MetricsFilter
	if strings.Trim(metricsFilter, " ") == "" {
		metricsFilter = ".*"
	}
	var filterMetricsName []string
	for _, metricName := range WorkspaceMetricsNames {
		bol, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && bol {
			filterMetricsName = append(filterMetricsName, metricName)
		}
	}

	var wgAll sync.WaitGroup
	var wsAllch = make(chan *[]FormatedMetric, ChannelMaxCapacityWorkspaceMetric)

	wsMap := getAllWorkspaces()

	for ws := range wsMap {
		// Only execute Prometheus queries for specific metrics on specific workspaces
		bol, err := regexp.MatchString(monitoringRequest.ResourcesFilter, ws)
		if err == nil && bol {
			// a workspace
			wgAll.Add(1)
			go collectWorkspaceMetric(monitoringRequest, ws, filterMetricsName, &wgAll, wsAllch)
		}
	}

	wgAll.Wait()
	close(wsAllch)

	fmtMetricMap := make(map[string]FormatedMetric)
	for oneWsMetric := range wsAllch {
		if oneWsMetric != nil {
			// aggregate workspace metric
			for _, metric := range *oneWsMetric {
				fm, exist := fmtMetricMap[metric.MetricName]
				if exist {
					if metric.Status == "error" {
						fm.Status = metric.Status
					}
					fm.Data.Result = append(fm.Data.Result, metric.Data.Result...)
					fmtMetricMap[metric.MetricName] = fm
				} else {
					fmtMetricMap[metric.MetricName] = metric
				}
			}
		}
	}

	var metricArray = make([]FormatedMetric, 0)
	for _, metric := range fmtMetricMap {
		metricArray = append(metricArray, metric)
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelClusterWorkspace,
		Results:      metricArray,
	}
}

func collectWorkspaceMetric(monitoringRequest *MonitoringRequestParams, ws string, filterMetricsName []string, wgAll *sync.WaitGroup, wsAllch chan *[]FormatedMetric) {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return
	}

	defer wgAll.Done()
	var wg sync.WaitGroup
	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	namespaceArray, err := workspaces.WorkspaceNamespaces(ws)
	if err != nil {
		klog.Errorln(err)
	}

	// add by namespace
	for _, metricName := range filterMetricsName {
		wg.Add(1)
		go func(metricName string) {

			queryType, params := AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest, namespaceArray, ws, metricName)
			metricsStr := client.SendMonitoringRequest(queryType, params)
			ch <- ReformatJson(metricsStr, metricName, map[string]string{ResultItemMetricResourceName: ws})
			wg.Done()
		}(metricName)
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric
	for oneMetric := range ch {
		if oneMetric != nil {
			// add "workspace" to oneMetric "metric" field
			for i := 0; i < len(oneMetric.Data.Result); i++ {
				tmap, sure := oneMetric.Data.Result[i][ResultItemMetric].(map[string]interface{})
				if sure {
					tmap[MetricLevelWorkspace] = ws
					oneMetric.Data.Result[i][ResultItemMetric] = tmap
				}
			}
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	wsAllch <- &metricsArray
}

func GetClusterLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range ClusterMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params := AssembleClusterMetricRequestInfo(monitoringRequest, metricName)
				metricsStr := client.SendMonitoringRequest(queryType, params)
				ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelCluster: "local"})
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelCluster,
		Results:      metricsArray,
	}
}

func GetNodeLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range NodeMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params := AssembleNodeMetricRequestInfo(monitoringRequest, metricName)
				metricsStr := client.SendMonitoringRequest(queryType, params)
				ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelNode: ""})
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelNode,
		Results:      metricsArray,
	}
}

func GetWorkspaceLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	// a specific workspace's metrics
	if monitoringRequest.WsName != "" {
		namespaceArray, err := workspaces.WorkspaceNamespaces(monitoringRequest.WsName)
		if err != nil {
			klog.Errorln(err.Error())
		}
		namespaceArray = filterNamespace(monitoringRequest.ResourcesFilter, namespaceArray)

		if monitoringRequest.Tp == "rank" {
			for _, metricName := range NamespaceMetricsNames {
				if metricName == MetricNameWorkspaceAllProjectCount {
					continue
				}

				matched, err := regexp.MatchString(metricsFilter, metricName)
				if err != nil || !matched {
					continue
				}

				wg.Add(1)
				go func(metricName string) {

					var chForOneMetric = make(chan *FormatedMetric, ChannelMaxCapacity)
					var wgForOneMetric sync.WaitGroup

					for _, ns := range namespaceArray {
						wgForOneMetric.Add(1)
						go func(metricName string, namespace string) {

							queryType, params := AssembleNamespaceMetricRequestInfoByNamesapce(monitoringRequest, namespace, metricName)
							metricsStr := client.SendMonitoringRequest(queryType, params)
							chForOneMetric <- ReformatJson(metricsStr, metricName, map[string]string{ResultItemMetricResourceName: namespace})
							wgForOneMetric.Done()
						}(metricName, ns)
					}

					wgForOneMetric.Wait()
					close(chForOneMetric)

					// ranking is for vector type result only
					aggregatedResult := FormatedMetric{MetricName: metricName, Status: MetricStatusSuccess, Data: FormatedMetricData{Result: []map[string]interface{}{}, ResultType: ResultTypeVector}}

					for oneMetric := range chForOneMetric {

						if oneMetric != nil {

							// append .data.result[0]
							if len(oneMetric.Data.Result) > 0 {
								aggregatedResult.Data.Result = append(aggregatedResult.Data.Result, oneMetric.Data.Result[0])
							}
						}
					}

					ch <- &aggregatedResult
					wg.Done()
				}(metricName)

			}

		} else {

			workspace := monitoringRequest.WsName

			for _, metricName := range WorkspaceMetricsNames {

				if metricName == MetricNameWorkspaceAllProjectCount {
					continue
				}

				matched, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && matched {
					wg.Add(1)
					go func(metricName string, workspace string) {
						queryType, params := AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest, namespaceArray, workspace, metricName)
						metricsStr := client.SendMonitoringRequest(queryType, params)
						ch <- ReformatJson(metricsStr, metricName, map[string]string{ResultItemMetricResourceName: workspace})
						wg.Done()
					}(metricName, workspace)
				}
			}
		}
	} else {
		// sum all workspaces
		for _, metricName := range WorkspaceMetricsNames {
			matched, err := regexp.MatchString(metricsFilter, metricName)
			if err == nil && matched {

				wg.Add(1)

				go func(metricName string) {
					queryType, params := AssembleAllWorkspaceMetricRequestInfo(monitoringRequest, nil, metricName)
					metricsStr := client.SendMonitoringRequest(queryType, params)
					ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelWorkspace: "workspaces"})

					wg.Done()
				}(metricName)
			}
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelWorkspace,
		Results:      metricsArray,
	}
}

func GetNamespaceLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range NamespaceMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {

				queryType, params := AssembleNamespaceMetricRequestInfo(monitoringRequest, metricName)
				metricsStr := client.SendMonitoringRequest(queryType, params)

				rawResult := ReformatJson(metricsStr, metricName, map[string]string{MetricLevelNamespace: ""})
				ch <- rawResult

				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelNamespace,
		Results:      metricsArray,
	}
}

func GetWorkloadLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	if monitoringRequest.WorkloadName == "" {
		for _, metricName := range WorkloadMetricsNames {
			matched, err := regexp.MatchString(metricsFilter, metricName)
			if err == nil && matched {
				wg.Add(1)
				go func(metricName string) {
					queryType, params := AssembleAllWorkloadMetricRequestInfo(monitoringRequest, metricName)
					metricsStr := client.SendMonitoringRequest(queryType, params)
					reformattedResult := ReformatJson(metricsStr, metricName, map[string]string{MetricLevelWorkload: ""})
					// no need to append a null result
					ch <- reformattedResult
					wg.Done()
				}(metricName)
			}
		}
	} else {
		for _, metricName := range WorkloadMetricsNames {
			bol, err := regexp.MatchString(metricsFilter, metricName)
			if err == nil && bol {
				wg.Add(1)
				go func(metricName string) {
					metricName = strings.TrimLeft(metricName, "workload_")
					queryType, params, nullRule := AssembleSpecificWorkloadMetricRequestInfo(monitoringRequest, metricName)
					if !nullRule {
						metricsStr := client.SendMonitoringRequest(queryType, params)
						fmtMetrics := ReformatJson(metricsStr, metricName, map[string]string{MetricLevelPodName: ""})
						unifyMetricHistoryTimeRange(fmtMetrics)
						ch <- fmtMetrics
					}
					wg.Done()
				}(metricName)
			}
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelWorkload,
		Results:      metricsArray,
	}
}

func GetPodLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range PodMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params, nullRule := AssemblePodMetricRequestInfo(monitoringRequest, metricName)
				if !nullRule {
					metricsStr := client.SendMonitoringRequest(queryType, params)
					ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelPodName: ""})
				} else {
					ch <- nil
				}
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelPod,
		Results:      metricsArray,
	}
}

func GetContainerLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range ContainerMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params := AssembleContainerMetricRequestInfo(monitoringRequest, metricName)
				metricsStr := client.SendMonitoringRequest(queryType, params)
				ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelContainerName: ""})
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelContainer,
		Results:      metricsArray,
	}
}

func GetPVCLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range PVCMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params, nullRule := AssemblePVCMetricRequestInfo(monitoringRequest, metricName)
				if !nullRule {
					metricsStr := client.SendMonitoringRequest(queryType, params)
					ch <- ReformatJson(metricsStr, metricName, map[string]string{MetricLevelPVC: ""})
				} else {
					ch <- nil
				}
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelPVC,
		Results:      metricsArray,
	}
}

func GetComponentLevelMetrics(monitoringRequest *MonitoringRequestParams) *FormatedLevelMetric {
	client, err := cs.ClientSets().Prometheus()
	if err != nil {
		klog.Error(err)
		return nil
	}

	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	for _, metricName := range ComponentMetricsNames {
		matched, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && matched {
			wg.Add(1)
			go func(metricName string) {
				queryType, params := AssembleComponentRequestInfo(monitoringRequest, metricName)
				metricsStr := client.SendMonitoringRequest(queryType, params)
				formattedJson := ReformatJson(metricsStr, metricName, map[string]string{ResultItemMetricResourceName: monitoringRequest.ComponentName})

				if metricName == EtcdServerList {

					nodeMap := make(map[string]string, 0)

					nodeAddress := GetNodeAddressInfo()
					for nodeName, nodeInfo := range *nodeAddress {

						var nodeIp string
						for _, item := range nodeInfo {
							if item.Type == v1.NodeInternalIP {
								nodeIp = item.Address
								break
							}
						}

						nodeMap[nodeIp] = nodeName
					}

					// add node_name label to metrics
					for i := 0; i < len(formattedJson.Data.Result); i++ {
						metricDesc := formattedJson.Data.Result[i][ResultItemMetric]
						metricDescMap, ensure := metricDesc.(map[string]interface{})
						if ensure {
							if nodeIp, exist := metricDescMap[ResultItemMetricNodeIp]; exist {
								metricDescMap[ResultItemMetricNodeName] = nodeMap[nodeIp.(string)]
							}
						}
					}
				}

				ch <- formattedJson
				wg.Done()
			}(metricName)
		}
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric

	for oneMetric := range ch {
		if oneMetric != nil {
			metricsArray = append(metricsArray, *oneMetric)
		}
	}

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelComponent,
		Results:      metricsArray,
	}
}

func GetAllWorkspacesStatistics() *FormatedLevelMetric {

	wg := sync.WaitGroup{}
	var metricsArray []FormatedMetric
	timestamp := time.Now().Unix()

	var orgResultItem *FormatedMetric
	var devopsResultItem *FormatedMetric
	var workspaceProjResultItem *FormatedMetric
	var accountResultItem *FormatedMetric

	wg.Add(4)

	go func() {
		orgNums, errOrg := workspaces.WorkspaceCount()
		if errOrg != nil {
			klog.Errorln(errOrg.Error())
		}
		orgResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllOrganizationCount, WorkspaceResourceKindOrganization, orgNums, errOrg)
		wg.Done()
	}()

	go func() {
		devOpsProjectNums, errDevops := workspaces.GetAllDevOpsProjectsNums()
		if errDevops != nil {
			klog.Errorln(errDevops.Error())
		}
		devopsResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllDevopsCount, WorkspaceResourceKindDevops, devOpsProjectNums, errDevops)
		wg.Done()
	}()

	go func() {
		projNums, errProj := workspaces.GetAllProjectNums()
		if errProj != nil {
			klog.Errorln(errProj.Error())
		}
		workspaceProjResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllProjectCount, WorkspaceResourceKindNamespace, projNums, errProj)
		wg.Done()
	}()

	go func() {
		result, errAct := cs.ClientSets().KubeSphere().ListUsers()
		if errAct != nil {
			klog.Errorln(errAct.Error())
		}
		accountResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllAccountCount, WorkspaceResourceKindAccount, result.TotalCount, errAct)
		wg.Done()
	}()

	wg.Wait()

	metricsArray = append(metricsArray, *orgResultItem, *devopsResultItem, *workspaceProjResultItem, *accountResultItem)

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelWorkspace,
		Results:      metricsArray,
	}
}

func MonitorOneWorkspaceStatistics(wsName string) *FormatedLevelMetric {

	var nsMetrics *FormatedMetric
	var devopsMetrics *FormatedMetric
	var memberMetrics *FormatedMetric
	var roleMetrics *FormatedMetric

	wg := sync.WaitGroup{}
	wg.Add(4)

	var fMetricsArray []FormatedMetric
	timestamp := int64(time.Now().Unix())

	go func() {
		// add namespaces(project) metric
		namespaces, errNs := workspaces.WorkspaceNamespaces(wsName)

		if errNs != nil {
			klog.Errorln(errNs.Error())
		}
		nsMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceNamespaceCount, WorkspaceResourceKindNamespace, len(namespaces), errNs)
		wg.Done()
	}()

	go func() {
		devOpsProjects, errDevOps := workspaces.GetDevOpsProjects(wsName)
		if errDevOps != nil {
			klog.Errorln(errDevOps.Error())
		}
		// add devops metric
		devopsMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceDevopsCount, WorkspaceResourceKindDevops, len(devOpsProjects), errDevOps)
		wg.Done()
	}()

	go func() {
		count, errMemb := workspaces.WorkspaceUserCount(wsName)
		if errMemb != nil {
			klog.Errorln(errMemb.Error())
		}
		// add member metric
		memberMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceMemberCount, WorkspaceResourceKindMember, count, errMemb)
		wg.Done()
	}()

	go func() {
		roles, errRole := workspaces.GetOrgRoles(wsName)
		if errRole != nil {
			klog.Errorln(errRole.Error())
		}
		// add role metric
		roleMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceRoleCount, WorkspaceResourceKindRole, len(roles), errRole)
		wg.Done()
	}()

	wg.Wait()

	fMetricsArray = append(fMetricsArray, *nsMetrics, *devopsMetrics, *memberMetrics, *roleMetrics)

	return &FormatedLevelMetric{
		MetricsLevel: MetricLevelWorkspace,
		Results:      fMetricsArray,
	}
}

func getSpecificMetricItem(timestamp int64, metricName string, resource string, count int, err error, resourceType ...string) *FormatedMetric {
	var nsMetrics FormatedMetric
	nsMetrics.MetricName = metricName
	nsMetrics.Data.ResultType = ResultTypeVector
	resultItem := make(map[string]interface{})
	tmp := make(map[string]string)

	if len(resourceType) > 0 {
		tmp[resourceType[0]] = resource
	} else {
		tmp[ResultItemMetricResource] = resource
	}

	if err == nil {
		nsMetrics.Status = MetricStatusSuccess
	} else {
		nsMetrics.Status = MetricStatusError
		resultItem["errormsg"] = err.Error()
	}

	resultItem[ResultItemMetric] = tmp
	resultItem[ResultItemValue] = []interface{}{timestamp, count}
	nsMetrics.Data.Result = make([]map[string]interface{}, 1)
	nsMetrics.Data.Result[0] = resultItem
	return &nsMetrics
}

func AssembleClusterMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	rule := MakeClusterRule(metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleNodeMetricRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	rule := MakeNodeRule(monitoringRequest.NodeId, monitoringRequest.ResourcesFilter, metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleComponentRequestInfo(monitoringRequest *MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	rule := MakeComponentRule(metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

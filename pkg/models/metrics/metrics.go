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
	"kubesphere.io/kubesphere/pkg/models"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/models/components"
	"kubesphere.io/kubesphere/pkg/models/workspaces"

	"github.com/golang/glog"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"runtime/debug"
	"sort"

	"github.com/json-iterator/go"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"

	"kubesphere.io/kubesphere/pkg/client"
)

var (
	jsonIter           = jsoniter.ConfigCompatibleWithStandardLibrary
	nodeStatusDelLabel = []string{"endpoint", "instance", "job", "namespace", "pod", "service"}
)

const (
	ChannelMaxCapacityWorkspaceMetric = 800
	ChannelMaxCapacity                = 100
)

type FormatedLevelMetric struct {
	MetricsLevel string           `json:"metrics_level"`
	Results      []FormatedMetric `json:"results"`
}

type FormatedMetric struct {
	MetricName string             `json:"metric_name,omitempty"`
	Status     string             `json:"status"`
	Data       FormatedMetricData `json:"data,omitempty"`
}

type FormatedMetricData struct {
	Result     []map[string]interface{} `json:"result"`
	ResultType string                   `json:"resultType"`
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

	paramValues := make(url.Values)
	paramValues.Set("query", WorkspaceNamespaceLabelRule)
	params := paramValues.Encode()
	res := client.SendMonitoringRequest(client.DefaultQueryType, params)

	metric := ReformatJson(res, "")

	return getAllWorkspaceNames(metric)
}

func getPodNameRegexInWorkload(res, filter string) string {

	data := []byte(res)
	var dat CommonMetricsResult
	jsonErr := jsonIter.Unmarshal(data, &dat)
	if jsonErr != nil {
		glog.Errorln("json parse failed", jsonErr.Error(), res)
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
			glog.Errorln(err)
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

func AssembleSpecificWorkloadMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string, bool) {

	nsName := monitoringRequest.NsName
	wkName := monitoringRequest.WorkloadName
	podsFilter := monitoringRequest.PodsFilter

	rule := MakeSpecificWorkloadRule(monitoringRequest.WorkloadKind, wkName, nsName)
	paramValues := monitoringRequest.Params
	params := makeRequestParamString(rule, paramValues)

	res := client.SendMonitoringRequest(client.DefaultQueryType, params)

	podNamesFilter := getPodNameRegexInWorkload(res, podsFilter)

	queryType := monitoringRequest.QueryType
	rule = MakePodPromQL(metricName, nsName, "", "", podNamesFilter)
	params = makeRequestParamString(rule, paramValues)

	return queryType, params, rule == ""
}

func AssembleAllWorkloadMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params

	rule := MakeWorkloadPromQL(metricName, monitoringRequest.NsName, monitoringRequest.WlFilter)
	params := makeRequestParamString(rule, paramValues)
	return queryType, params
}

func AssemblePodMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string, bool) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params

	rule := MakePodPromQL(metricName, monitoringRequest.NsName, monitoringRequest.NodeId, monitoringRequest.PodName, monitoringRequest.PodsFilter)
	params := makeRequestParamString(rule, paramValues)
	return queryType, params, rule == ""
}

func GetMetric(queryType, params, metricName string) *FormatedMetric {
	res := client.SendMonitoringRequest(queryType, params)
	formatedMetric := ReformatJson(res, metricName)
	return formatedMetric
}

func GetNodeAddressInfo() *map[string][]v1.NodeAddress {
	nodeLister := informers.SharedInformerFactory().Core().V1().Nodes().Lister()
	nodes, err := nodeLister.List(labels.Everything())

	if err != nil {
		glog.Errorln(err.Error())
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
			if nodeId, exist := metricDescMap["node"]; exist {
				addr, exist := (*nodeAddress)[nodeId.(string)]
				if exist {
					metricDescMap["address"] = addr
				}
			}
		}
	}
}

func MonitorContainer(monitoringRequest *client.MonitoringRequestParams, metricName string) *FormatedMetric {
	queryType, params := AssembleContainerMetricRequestInfo(monitoringRequest, metricName)
	res := GetMetric(queryType, params, metricName)
	return res
}

func AssembleContainerMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params
	rule := MakeContainerPromQL(monitoringRequest.NsName, monitoringRequest.NodeId, monitoringRequest.PodName, monitoringRequest.ContainerName, metricName, monitoringRequest.ContainersFilter)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleNamespaceMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType

	paramValues := monitoringRequest.Params
	rule := MakeNamespacePromQL(monitoringRequest.NsName, monitoringRequest.NsFilter, metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, namespaceList []string, metricName string) (string, string) {

	nsFilter := "^(" + strings.Join(namespaceList, "|") + ")$"

	queryType := monitoringRequest.QueryType

	rule := MakeSpecificWorkspacePromQL(metricName, nsFilter)
	paramValues := monitoringRequest.Params
	params := makeRequestParamString(rule, paramValues)
	return queryType, params
}

func AssembleAllWorkspaceMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, namespaceList []string, metricName string) (string, string) {
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
			glog.Errorln(err)
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

func MonitorAllWorkspaces(monitoringRequest *client.MonitoringRequestParams) *FormatedLevelMetric {
	metricsFilter := monitoringRequest.MetricsFilter
	if strings.Trim(metricsFilter, " ") == "" {
		metricsFilter = ".*"
	}
	var filterMetricsName []string
	for _, metricName := range WorkspaceMetricsNames {
		if metricName == MetricNameWorkspaceAllProjectCount {
			continue
		}
		bol, err := regexp.MatchString(metricsFilter, metricName)
		if err == nil && bol {
			filterMetricsName = append(filterMetricsName, metricName)
		}
	}

	var wgAll sync.WaitGroup
	var wsAllch = make(chan *[]FormatedMetric, ChannelMaxCapacityWorkspaceMetric)

	wsMap := getAllWorkspaces()

	for ws := range wsMap {
		bol, err := regexp.MatchString(monitoringRequest.WsFilter, ws)
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

func collectWorkspaceMetric(monitoringRequest *client.MonitoringRequestParams, ws string, filterMetricsName []string, wgAll *sync.WaitGroup, wsAllch chan *[]FormatedMetric) {
	defer wgAll.Done()
	var wg sync.WaitGroup
	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	namespaceArray, err := workspaces.WorkspaceNamespaces(ws)
	if err != nil {
		glog.Errorln(err.Error())
	}
	// add by namespace
	for _, metricName := range filterMetricsName {
		wg.Add(1)
		go func(metricName string) {

			queryType, params := AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest, namespaceArray, metricName)
			ch <- GetMetric(queryType, params, metricName)

			wg.Done()
		}(metricName)
	}

	wg.Wait()
	close(ch)

	var metricsArray []FormatedMetric
	for oneMetric := range ch {
		if oneMetric != nil {
			// add "workspace" filed to oneMetric `metric` field
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

func MonitorAllMetrics(monitoringRequest *client.MonitoringRequestParams, resourceType string) *FormatedLevelMetric {
	metricsFilter := monitoringRequest.MetricsFilter
	if metricsFilter == "" {
		metricsFilter = ".*"
	}

	var ch = make(chan *FormatedMetric, ChannelMaxCapacity)
	var wg sync.WaitGroup

	switch resourceType {
	case MetricLevelCluster:
		{
			for _, metricName := range ClusterMetricsNames {
				bol, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && bol {
					wg.Add(1)
					go func(metricName string) {
						queryType, params := AssembleClusterMetricRequestInfo(monitoringRequest, metricName)
						clusterMetrics := GetMetric(queryType, params, metricName)

						ch <- clusterMetrics

						wg.Done()
					}(metricName)
				}
			}
		}
	case MetricLevelNode:
		{
			for _, metricName := range NodeMetricsNames {
				bol, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && bol {
					wg.Add(1)
					go func(metricName string) {
						queryType, params := AssembleNodeMetricRequestInfo(monitoringRequest, metricName)
						ch <- GetMetric(queryType, params, metricName)
						wg.Done()
					}(metricName)
				}
			}
		}
	case MetricLevelWorkspace:
		{
			// a specific workspace's metrics
			if monitoringRequest.WsName != "" {
				namespaceArray, err := workspaces.WorkspaceNamespaces(monitoringRequest.WsName)
				if err != nil {
					glog.Errorln(err.Error())
				}
				namespaceArray = filterNamespace(monitoringRequest.NsFilter, namespaceArray)

				if monitoringRequest.Tp == "rank" {
					for _, metricName := range NamespaceMetricsNames {
						if metricName == MetricNameWorkspaceAllProjectCount {
							continue
						}

						bol, err := regexp.MatchString(metricsFilter, metricName)
						ns := "^(" + strings.Join(namespaceArray, "|") + ")$"
						monitoringRequest.NsFilter = ns
						if err == nil && bol {
							wg.Add(1)
							go func(metricName string) {
								queryType, params := AssembleNamespaceMetricRequestInfo(monitoringRequest, metricName)
								ch <- GetMetric(queryType, params, metricName)
								wg.Done()
							}(metricName)
						}
					}

				} else {
					for _, metricName := range WorkspaceMetricsNames {

						if metricName == MetricNameWorkspaceAllProjectCount {
							continue
						}

						bol, err := regexp.MatchString(metricsFilter, metricName)
						if err == nil && bol {
							wg.Add(1)
							go func(metricName string) {
								queryType, params := AssembleSpecificWorkspaceMetricRequestInfo(monitoringRequest, namespaceArray, metricName)
								ch <- GetMetric(queryType, params, metricName)
								wg.Done()
							}(metricName)
						}
					}
				}
			} else {
				// sum all workspaces

				for _, metricName := range WorkspaceMetricsNames {
					bol, err := regexp.MatchString(metricsFilter, metricName)
					if err == nil && bol {

						wg.Add(1)

						go func(metricName string) {
							queryType, params := AssembleAllWorkspaceMetricRequestInfo(monitoringRequest, nil, metricName)

							ch <- GetMetric(queryType, params, metricName)

							wg.Done()
						}(metricName)
					}
				}
			}
		}
	case MetricLevelNamespace:
		{
			for _, metricName := range NamespaceMetricsNames {
				bol, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && bol {
					wg.Add(1)
					go func(metricName string) {
						queryType, params := AssembleNamespaceMetricRequestInfo(monitoringRequest, metricName)
						ch <- GetMetric(queryType, params, metricName)
						wg.Done()
					}(metricName)
				}
			}
		}
	case MetricLevelWorkload:
		{
			if monitoringRequest.Tp == "rank" {
				for _, metricName := range WorkloadMetricsNames {
					bol, err := regexp.MatchString(metricsFilter, metricName)
					if err == nil && bol {
						wg.Add(1)
						go func(metricName string) {
							queryType, params := AssembleAllWorkloadMetricRequestInfo(monitoringRequest, metricName)
							fmtMetrics := GetMetric(queryType, params, metricName)
							ch <- fmtMetrics
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
								fmtMetrics := GetMetric(queryType, params, metricName)
								unifyMetricHistoryTimeRange(fmtMetrics)
								ch <- fmtMetrics
							}
							wg.Done()
						}(metricName)
					}
				}
			}
		}
	case MetricLevelPod:
		{
			for _, metricName := range PodMetricsNames {
				bol, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && bol {
					wg.Add(1)
					go func(metricName string) {
						queryType, params, nullRule := AssemblePodMetricRequestInfo(monitoringRequest, metricName)
						if !nullRule {
							ch <- GetMetric(queryType, params, metricName)
						} else {
							ch <- nil
						}
						wg.Done()
					}(metricName)
				}
			}
		}
	case MetricLevelContainer:
		{
			for _, metricName := range ContainerMetricsNames {
				bol, err := regexp.MatchString(metricsFilter, metricName)
				if err == nil && bol {
					wg.Add(1)
					go func(metricName string) {
						queryType, params := AssembleContainerMetricRequestInfo(monitoringRequest, metricName)
						ch <- GetMetric(queryType, params, metricName)
						wg.Done()
					}(metricName)
				}
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
		MetricsLevel: resourceType,
		Results:      metricsArray,
	}
}

func MonitorAllWorkspacesStatistics() *FormatedLevelMetric {

	wg := sync.WaitGroup{}
	var metricsArray []FormatedMetric
	timestamp := time.Now().Unix()

	var orgResultItem *FormatedMetric
	var devopsResultItem *FormatedMetric
	var workspaceProjResultItem *FormatedMetric
	var accountResultItem *FormatedMetric

	wg.Add(4)

	go func() {
		orgNums, errOrg := workspaces.Count()
		if errOrg != nil {
			glog.Errorln(errOrg.Error())
		}
		orgResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllOrganizationCount, WorkspaceResourceKindOrganization, orgNums, errOrg)
		wg.Done()
	}()

	go func() {
		devOpsProjectNums, errDevops := workspaces.GetAllDevOpsProjectsNums()
		if errDevops != nil {
			glog.Errorln(errDevops.Error())
		}
		devopsResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllDevopsCount, WorkspaceResourceKindDevops, devOpsProjectNums, errDevops)
		wg.Done()
	}()

	go func() {
		projNums, errProj := workspaces.GetAllProjectNums()
		if errProj != nil {
			glog.Errorln(errProj.Error())
		}
		workspaceProjResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllProjectCount, WorkspaceResourceKindNamespace, projNums, errProj)
		wg.Done()
	}()

	go func() {
		actNums, errAct := workspaces.GetAllAccountNums()
		if errAct != nil {
			glog.Errorln(errAct.Error())
		}
		accountResultItem = getSpecificMetricItem(timestamp, MetricNameWorkspaceAllAccountCount, WorkspaceResourceKindAccount, actNums, errAct)
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
			glog.Errorln(errNs.Error())
		}
		nsMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceNamespaceCount, WorkspaceResourceKindNamespace, len(namespaces), errNs)
		wg.Done()
	}()

	go func() {
		devOpsProjects, errDevOps := workspaces.GetDevOpsProjects(wsName)
		if errDevOps != nil {
			glog.Errorln(errDevOps.Error())
		}
		// add devops metric
		devopsMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceDevopsCount, WorkspaceResourceKindDevops, len(devOpsProjects), errDevOps)
		wg.Done()
	}()

	go func() {
		members, errMemb := workspaces.GetOrgMembers(wsName)
		if errMemb != nil {
			glog.Errorln(errMemb.Error())
		}
		// add member metric
		memberMetrics = getSpecificMetricItem(timestamp, MetricNameWorkspaceMemberCount, WorkspaceResourceKindMember, len(members), errMemb)
		wg.Done()
	}()

	go func() {
		roles, errRole := workspaces.GetOrgRoles(wsName)
		if errRole != nil {
			glog.Errorln(errRole.Error())
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

// k8s component(controller, scheduler, etcd) status
func MonitorComponentStatus(monitoringRequest *client.MonitoringRequestParams) *[]interface{} {
	componentList, err := client.K8sClient().CoreV1().ComponentStatuses().List(metaV1.ListOptions{})
	if err != nil {
		glog.Errorln(err.Error())
	}

	var componentStatusList []*ComponentStatus
	for _, item := range componentList.Items {
		var status []OneComponentStatus
		for _, cond := range item.Conditions {
			status = append(status, OneComponentStatus{
				Type:    string(cond.Type),
				Status:  string(cond.Status),
				Message: cond.Message,
				Error:   cond.Error,
			})
		}

		componentStatusList = append(componentStatusList, &ComponentStatus{
			Name:            item.Name,
			Namespace:       item.Namespace,
			Labels:          item.Labels,
			ComponentStatus: status,
		})
	}

	// node status
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	paramValues.Set("query", NodeStatusRule)
	params := paramValues.Encode()
	res := client.SendMonitoringRequest(queryType, params)

	nodeStatusMetric := ReformatJson(res, "node_status", nodeStatusDelLabel...)
	nodeStatusMetric = ReformatNodeStatusField(nodeStatusMetric)

	var normalNodes []string
	var abnormalNodes []string
	for _, result := range nodeStatusMetric.Data.Result {
		tmap, sure := result[ResultItemMetric].(map[string]interface{})

		if sure {
			if tmap[MetricStatus].(string) == "false" {
				abnormalNodes = append(abnormalNodes, tmap[MetricLevelNode].(string))
			} else {
				normalNodes = append(normalNodes, tmap[MetricLevelNode].(string))
			}
		}
	}

	Components, err := components.GetAllComponentsStatus()

	if err != nil {
		glog.Error(err.Error())
	}

	var namespaceComponentHealthyMap = make(map[string]int)
	var namespaceComponentTotalMap = make(map[string]int)

	for _, ns := range constants.SystemNamespaces {
		nsStatus, exist := Components[ns]
		if exist {
			for _, nsStatusItem := range nsStatus.(map[string]interface{}) {
				component := nsStatusItem.(models.Component)
				namespaceComponentTotalMap[ns] += 1
				if component.HealthyBackends != 0 && component.HealthyBackends == component.TotalBackends {
					namespaceComponentHealthyMap[ns] += 1
				}
			}
		}
	}

	timestamp := int64(time.Now().Unix())

	onlineMetricItems := makeMetricItems(timestamp, namespaceComponentHealthyMap, MetricLevelNamespace)
	metricItems := makeMetricItems(timestamp, namespaceComponentTotalMap, MetricLevelNamespace)

	var assembleList []interface{}
	assembleList = append(assembleList, nodeStatusMetric)

	for _, statusItem := range componentStatusList {
		assembleList = append(assembleList, statusItem)
	}

	assembleList = append(assembleList, FormatedMetric{
		Data: FormatedMetricData{
			Result:     *onlineMetricItems,
			ResultType: ResultTypeVector,
		},
		MetricName: MetricNameComponentOnLine,
		Status:     MetricStatusSuccess,
	})

	assembleList = append(assembleList, FormatedMetric{
		Data: FormatedMetricData{
			Result:     *metricItems,
			ResultType: ResultTypeVector,
		},
		MetricName: MetricNameComponentLine,
		Status:     MetricStatusSuccess,
	})

	return &assembleList
}

func makeMetricItems(timestamp int64, statusMap map[string]int, resourceType string) *[]map[string]interface{} {
	var metricItems []map[string]interface{}

	for ns, count := range statusMap {
		metricItems = append(metricItems, map[string]interface{}{
			ResultItemMetric: map[string]string{resourceType: ns},
			ResultItemValue:  []interface{}{timestamp, fmt.Sprintf("%d", count)},
		})
	}
	return &metricItems
}

func AssembleClusterMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	rule := MakeClusterRule(metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

func AssembleNodeMetricRequestInfo(monitoringRequest *client.MonitoringRequestParams, metricName string) (string, string) {
	queryType := monitoringRequest.QueryType
	paramValues := monitoringRequest.Params
	rule := MakeNodeRule(monitoringRequest.NodeId, monitoringRequest.NodesFilter, metricName)
	params := makeRequestParamString(rule, paramValues)

	return queryType, params
}

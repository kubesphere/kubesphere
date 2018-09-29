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

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"strconv"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/organizations"
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
	if strings.Contains(path, "workloads") {
		sourceType = "workload"
	} else if strings.Contains(path, "monitoring/workspaces") {
		sourceType = "workspace"
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
			if sourceType == "node" || sourceType == "cluster" {
				go collectNodeorClusterMetrics(request, metricName, ch)
			} else if sourceType == "namespace" {
				go collectNamespaceMetrics(request, metricName, ch)
			} else if sourceType == "pod" {
				go collectPodMetrics(request, metricName, ch)
			} else if sourceType == "workload" {
				go collectWorkloadMetrics(request, metricName, ch)
			} else if sourceType == "workspace" {
				name := request.QueryParameter("workspace_name")
				namespaceArray, err := organizations.GetNamespaces(name)
				if err != nil {
					glog.Errorln(err)
				}
				namespaceArray = filterNamespace(request, namespaceArray)
				go collectWorkspaceMetrics(request, metricName, namespaceArray, ch)
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

func MonitorWorkspaceUserInfo(req *restful.Request) FormatedMetric {
	orgNums, err1 := organizations.GetAllOrgNums()
	devOpsProjectNums, err2 := organizations.GetAllDevOpsProjectsNums()
	memberNums, err3 := organizations.GetOrgMembersNums()
	roleNums, err4 := organizations.GetOrgRolesNums()

	var fMetric FormatedMetric
	fMetric.Data.ResultType = "vector"
	fMetric.MetricName = "workspace_user_info_count"
	fMetric.Status = "success"
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		fMetric.Status = "error"
	}
	timestamp := time.Now().Unix()
	orgResultItem := getWorkspaceInfoItem(timestamp, orgNums, "organizations")
	dvpResultItem := getWorkspaceInfoItem(timestamp, devOpsProjectNums, "accounts")
	membResultItem := getWorkspaceInfoItem(timestamp, memberNums, "projects")
	roleResultItem := getWorkspaceInfoItem(timestamp, roleNums, "devops_projects")
	var resultItems []map[string]interface{}
	resultItems = append(resultItems, orgResultItem, dvpResultItem, membResultItem, roleResultItem)
	fMetric.Data.Result = resultItems
	return fMetric
}

func getWorkspaceInfoItem(timestamp int64, namespaceNums int64, resourceName string) map[string]interface{} {
	resultItem := make(map[string]interface{})
	tmp := make(map[string]string)
	tmp["resource"] = resourceName
	resultItem["metric"] = tmp
	resultItem["value"] = []interface{}{timestamp, strconv.FormatInt(namespaceNums, 10)}
	return resultItem
}

func MonitorWorkspaceResourceLevelMetrics(request *restful.Request) FormatedLevelMetric {
	wsName := request.PathParameter("workspace_name")
	namespaces, errNs := organizations.GetNamespaces(wsName)
	devOpsProjects, errDevOps := organizations.GetDevOpsProjects(wsName)
	members, errMemb := organizations.GetOrgMembers(wsName)
	roles, errRole := organizations.GetOrgRoles(wsName)

	var fMetricsArray []FormatedMetric
	timestamp := int64(time.Now().Unix())
	namespaces, noneExistentNs := getExistingNamespace(namespaces)
	if len(noneExistentNs) != 0 {
		nsStr := strings.Join(noneExistentNs, "|")
		errStr := "the namespaces " + nsStr + " do not exist"
		if errNs == nil {
			errNs = errors.New(errStr)
		} else {
			errNs = errors.New(errNs.Error() + "\t" + errStr)
		}
	}

	// add namespaces(project) metric
	nsMetrics := getWorkspaceMetrics(timestamp, "workspace_namespaces_count", "namespace", len(namespaces), errNs)
	// add devops metric
	devopsMetrics := getWorkspaceMetrics(timestamp, "workspace_devops_projects_count", "devops", len(devOpsProjects), errDevOps)
	// add member metric
	memberMetrics := getWorkspaceMetrics(timestamp, "workspace_members_count", "member", len(members), errMemb)
	// add role metric
	roleMetrics := getWorkspaceMetrics(timestamp, "workspace_roles_count", "role", len(roles), errRole)
	// add workloads count metric
	wlMetrics := getWorkspaceWorkloadCountMetrics(namespaces)
	// add pods count metric
	podsCountMetrics := getWorkspacePodsCountMetrics(request, namespaces)
	fMetricsArray = append(fMetricsArray, nsMetrics, devopsMetrics, memberMetrics, roleMetrics, wlMetrics, *podsCountMetrics)

	return FormatedLevelMetric{
		MetricsLevel: "workspace",
		Results:      fMetricsArray,
	}
}

func getWorkspacePodsCountMetrics(request *restful.Request, namespaces []string) *FormatedMetric {
	metricName := "namespace_pod_count"
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
	for _, ns := range namespaces {
		quotaMetric, err := models.GetNamespaceQuota(ns)
		if err != nil {
			glog.Errorln(err)
			continue
		}
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
	}
	wlMetrics := convertQuota2MetricStruct(&wlQuotaMetrics)
	return wlMetrics
}

func getWorkspaceMetrics(timestamp int64, metricName string, kind string, count int, err error) FormatedMetric {
	var nsMetrics FormatedMetric
	nsMetrics.MetricName = metricName
	nsMetrics.Data.ResultType = "vector"
	resultItem := make(map[string]interface{})
	tmp := make(map[string]string)
	tmp["resource"] = kind
	if err == nil {
		nsMetrics.Status = "success"
		resultItem["metric"] = tmp
		resultItem["value"] = []interface{}{timestamp, count}
	} else {
		nsMetrics.Status = "error"
		resultItem["errorinfo"] = err.Error()
		resultItem["metric"] = tmp
		resultItem["value"] = []interface{}{timestamp, count}
	}
	nsMetrics.Data.Result = make([]map[string]interface{}, 1)
	nsMetrics.Data.Result[0] = resultItem
	return nsMetrics
}

func MonitorNodeorClusterSingleMetric(request *restful.Request, metricsName string) *FormatedMetric {
	recordingRule := MakeNodeorClusterRule(request, metricsName)
	res := client.SendPrometheusRequest(request, recordingRule)
	cleanedJson := ReformatJson(res, metricsName)
	return cleanedJson
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
		fMetric.Status = "error"
		fMetric.Data.ResultType = ""
		errInfo := make(map[string]interface{})
		if err != nil {
			errInfo["errorinfo"] = err.Error()
		} else {
			errInfo["errorinfo"] = "namespace " + namespace + " does not exist"
		}
		fMetric.Data.Result = []map[string]interface{}{errInfo}
	}

	return fMetric
}

func convertQuota2MetricStruct(quotaMetric *models.ResourceQuota) FormatedMetric {
	var fMetric FormatedMetric
	fMetric.MetricName = "workload_count"
	fMetric.Status = "success"
	fMetric.Data.ResultType = "vector"
	timestamp := int64(time.Now().Unix())
	var resultItems []map[string]interface{}

	hardMap := make(map[string]string)
	for resourceName, v := range quotaMetric.Data.Hard {
		hardMap[resourceName.String()] = v.String()
	}

	for resourceName, v := range quotaMetric.Data.Used {
		resultItem := make(map[string]interface{})
		tmp := make(map[string]string)
		tmp["resource"] = resourceName.String()
		resultItem["metric"] = tmp
		resultItem["value"] = []interface{}{timestamp, hardMap[resourceName.String()], v.String()}
		resultItems = append(resultItems, resultItem)
	}

	fMetric.Data.Result = resultItems
	return fMetric
}

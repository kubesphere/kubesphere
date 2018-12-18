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

package log

import (
	//"fmt"
	//"encoding/json"
	//"regexp"

	"strconv"
	"strings"

	"github.com/emicklei/go-restful"
	"github.com/golang/glog"

	//"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/constants"
	//"kubesphere.io/kubesphere/pkg/models"
	"github.com/olivere/elastic"
)

func queryLabel(label string, labels_query []string) bool {
	var result = false

	for _, label_query := range labels_query {
		if strings.Contains(label, label_query) {
			result = true
			break
		}
	}

	return result
}

//Input: workspace_query, multiple workspace query keyword
//Output: namespaces which workspace string contains query keyword
func queryWorkspace(workspace_query string) []string {
	if workspace_query == "" {
		return nil
	}

	nsList, err := client.NewK8sClient().CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return nil
	}

	var namespaces []string

	label_query := strings.ToLower(strings.Replace(workspace_query, ",", " ", -1))
	labels_query := strings.Split(label_query, " ")
	glog.Infof("labels_query %v", labels_query)

	for _, ns := range nsList.Items {
		labels := ns.GetLabels()
		_, ok := labels[constants.WorkspaceLabelKey]
		if ok {
			if queryLabel(strings.ToLower(labels[constants.WorkspaceLabelKey]), labels_query) {
				namespaces = append(namespaces, ns.GetName())
			}
		}
	}

	return namespaces
}

func getNamespacesFromWorkspace() {

}

func LogQuery(level constants.LogQueryLevel, request *restful.Request) *elastic.SearchResult {
	var param client.QueryParameters

	param.Level = level

	//TODO: Get Namespace info from user workspace namespace
	//      Get Pod info from workload
	switch level {
	case constants.QueryLevelCluster:
		{
			param.Namespaces = queryWorkspace(request.QueryParameter("workspace_query"))
			glog.Infof("queryWorkspace return %v", param.Namespaces)
			param.Namespace_query = request.QueryParameter("namespace_query")
			//param.Workloads_query = request.QueryParameter("workload_query")
			param.Pod_query = request.QueryParameter("pod_query")
			param.Container_query = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkspace:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespace_query = request.QueryParameter("namespace_query")
			//param.Workloads_query = request.QueryParameter("workload_query")
			param.Pod_query = request.QueryParameter("pod_query")
			param.Container_query = request.QueryParameter("container_query")
		}
	case constants.QueryLevelNamespace:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads_query = request.QueryParameter("workload_query")
			param.Pod_query = request.QueryParameter("pod_query")
			param.Container_query = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkload:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pod_query = request.QueryParameter("pod_query")
			param.Container_query = request.QueryParameter("container_query")
		}
	case constants.QueryLevelPod:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods = []string{request.PathParameter("pod_name")}
			param.Container_query = request.QueryParameter("container_query")
		}
	case constants.QueryLevelContainer:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods = []string{request.PathParameter("pod_name")}
			param.Containers = []string{request.PathParameter("container_name")}
		}
	}

	param.Log_query = request.QueryParameter("log_query")
	param.Start_time = request.QueryParameter("start_time")
	param.End_time = request.QueryParameter("end_time")

	var err error
	param.From, err = strconv.Atoi(request.QueryParameter("from"))
	if err != nil {
		param.From = 0
	}
	param.Size, err = strconv.Atoi(request.QueryParameter("size"))
	if err != nil {
		param.Size = 10
	}

	//log.Printf("LogQuery with %v", param)

	glog.Infof("LogQuery with %v", param)

	return client.Query(param)
}

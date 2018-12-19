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

func intersection(s1, s2 []string) (inter []string) {
	hash := make(map[string]bool)
	for _, e := range s1 {
		hash[e] = true
	}
	for _, e := range s2 {
		// If elements present in the hashmap then append intersection list.
		if hash[e] {
			inter = append(inter, e)
		}
	}
	//Remove dups from slice.
	inter = removeDups(inter)
	return
}

//Remove dups from slice.
func removeDups(elements []string) (nodups []string) {
	encountered := make(map[string]bool)
	for _, element := range elements {
		if !encountered[element] {
			nodups = append(nodups, element)
			encountered[element] = true
		}
	}
	return
}

func matchLabel(label string, labelsMatch []string) bool {
	var result = false

	for _, labelMatch := range labelsMatch {
		if strings.Compare(label, labelMatch) == 0 {
			result = true
			break
		}
	}

	return result
}

func queryLabel(label string, labelsQuery []string) bool {
	var result = false

	for _, labelQuery := range labelsQuery {
		if strings.Contains(label, labelQuery) {
			result = true
			break
		}
	}

	return result
}

func queryWorkspace(workspaceMatch string, workspaceQuery string) (bool, []string) {
	if workspaceMatch == "" && workspaceQuery == "" {
		return false, nil
	}

	nsList, err := client.NewK8sClient().CoreV1().Namespaces().List(metaV1.ListOptions{})
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return true, nil
	}

	var namespaces []string

	var hasMatch = false
	var labelsMatch []string
	if workspaceMatch != "" {
		labelsMatch = strings.Split(strings.Replace(workspaceMatch, ",", " ", -1), " ")
		hasMatch = true
	}

	var hasQuery = false
	var labelsQuery []string
	if workspaceQuery != "" {
		labelsQuery = strings.Split(strings.ToLower(strings.Replace(workspaceQuery, ",", " ", -1)), " ")
		hasQuery = true
	}

	for _, ns := range nsList.Items {
		labels := ns.GetLabels()
		_, ok := labels[constants.WorkspaceLabelKey]
		if ok {
			var namespaceMatch = true
			if hasMatch {
				if !matchLabel(labels[constants.WorkspaceLabelKey], labelsMatch) {
					namespaceMatch = false
				}
			}
			if hasQuery {
				if !queryLabel(strings.ToLower(labels[constants.WorkspaceLabelKey]), labelsQuery) {
					namespaceMatch = false
				}
			}

			if namespaceMatch {
				namespaces = append(namespaces, ns.GetName())
			}
		}
	}

	return true, namespaces
}

func matchNamespace(namespaceMatch string, namespaceFilled bool, namespaces []string) (bool, []string) {
	if namespaceMatch == "" {
		return namespaceFilled, namespaces
	}

	namespacesMatch := strings.Split(strings.Replace(namespaceMatch, ",", " ", -1), " ")

	if namespaceFilled {
		return true, intersection(namespacesMatch, namespaces)
	}

	return true, namespacesMatch
}

func matchWorkload(workloadMatch string) (bool, []string) {
	return false, nil
}

func matchPod(podMatch string) (bool, []string) {
	if podMatch == "" {
		return false, nil
	}

	return true, strings.Split(strings.Replace(podMatch, ",", " ", -1), " ")
}

func matchContainer(containerMatch string) (bool, []string) {
	if containerMatch == "" {
		return false, nil
	}

	return true, strings.Split(strings.Replace(containerMatch, ",", " ", -1), " ")
}

func LogQuery(level constants.LogQueryLevel, request *restful.Request) *elastic.SearchResult {
	var param client.QueryParameters

	param.Level = level

	switch level {
	case constants.QueryLevelCluster:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.QueryParameter("workspaces"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			matchWorkload(request.QueryParameter("workloads"))
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"))
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchPod(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkspace:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			//param.Workloads_query = request.QueryParameter("workload_query")
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelNamespace:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads_query = request.QueryParameter("workload_query")
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkload:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelPod:
		{
			//param.Workspaces = strings.Split(request.PathParameter("workspace_name"), ",")
			param.Namespaces = []string{request.PathParameter("namespace_name")}
			//param.Workloads = strings.Split(request.PathParameter("workload_name"), ",")
			param.Pods = []string{request.PathParameter("pod_name")}
			param.ContainerQuery = request.QueryParameter("container_query")
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

	param.LogQuery = request.QueryParameter("log_query")
	param.StartTime = request.QueryParameter("start_time")
	param.EndTime = request.QueryParameter("end_time")

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

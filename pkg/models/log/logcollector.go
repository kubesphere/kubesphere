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
	"reflect"

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

func In(value interface{}, container interface{}) int {
	if container == nil {
		return -1
	}
	containerValue := reflect.ValueOf(container)
	switch reflect.TypeOf(container).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < containerValue.Len(); i++ {
			if containerValue.Index(i).Interface() == value {
				return i
			}
		}
	case reflect.Map:
		if containerValue.MapIndex(reflect.ValueOf(value)).IsValid() {
			return -1
		}
	default:
		return -1
	}
	return -1
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

func matchWorkload(workloadMatch string, namespaces []string) (bool, []string) {
	if workloadMatch == "" {
		return false, nil
	}

	workloadsMatch := strings.Split(strings.Replace(strings.Replace(workloadMatch, ",", " ", -1), "Deployment", "ReplicaSet", -1), " ")

	podList, err := client.NewK8sClient().CoreV1().Pods("").List(metaV1.ListOptions{})
	if err != nil {
		glog.Error("failed to list pods, error: ", err)
		return true, nil
	}

	var pods []string

	if namespaces == nil {
		for _, pod := range podList.Items {
			//glog.Infof("List Pod %v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			if len(pod.ObjectMeta.OwnerReferences) > 0 && In(pod.ObjectMeta.OwnerReferences[0].Kind, workloadsMatch) >= 0 {
				pods = append(pods, pod.Name)
			}
		}
	} else {
		for _, pod := range podList.Items {
			//glog.Infof("List Pod %v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			if len(pod.ObjectMeta.OwnerReferences) > 0 && In(pod.ObjectMeta.OwnerReferences[0].Kind, workloadsMatch) >= 0 && In(pod.Namespace, namespaces) >= 0 {
				pods = append(pods, pod.Name)
			}
		}
	}

	return true, pods
}

func matchPod(podMatch string, podFilled bool, pods []string) (bool, []string) {
	if podMatch == "" {
		return podFilled, pods
	}

	podsMatch := strings.Split(strings.Replace(podMatch, ",", " ", -1), " ")

	if podFilled {
		return true, intersection(podsMatch, pods)
	}

	return true, podsMatch
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
	param.Operation = request.QueryParameter("operation")

	switch level {
	case constants.QueryLevelCluster:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.QueryParameter("workspaces"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = matchWorkload(request.QueryParameter("workloads"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkspace:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace_name"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.QueryParameter("namespaces"), param.NamespaceFilled, param.Namespaces)
			param.NamespaceQuery = request.QueryParameter("namespace_query")
			param.PodFilled, param.Pods = matchWorkload(request.QueryParameter("workloads"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelNamespace:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace_name"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace_name"), param.NamespaceFilled, param.Namespaces)
			param.PodFilled, param.Pods = matchWorkload(request.QueryParameter("workloads"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelWorkload:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace_name"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace_name"), param.NamespaceFilled, param.Namespaces)
			param.PodFilled, param.Pods = matchWorkload(request.PathParameter("workload_name"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.QueryParameter("pods"), param.PodFilled, param.Pods)
			param.PodQuery = request.QueryParameter("pod_query")
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelPod:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace_name"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace_name"), param.NamespaceFilled, param.Namespaces)
			param.PodFilled, param.Pods = matchWorkload(request.PathParameter("workload_name"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.PathParameter("pod_name"), param.PodFilled, param.Pods)
			param.ContainerFilled, param.Containers = matchContainer(request.QueryParameter("containers"))
			param.ContainerQuery = request.QueryParameter("container_query")
		}
	case constants.QueryLevelContainer:
		{
			param.NamespaceFilled, param.Namespaces = queryWorkspace(request.PathParameter("workspace_name"), request.QueryParameter("workspace_query"))
			param.NamespaceFilled, param.Namespaces = matchNamespace(request.PathParameter("namespace_name"), param.NamespaceFilled, param.Namespaces)
			param.PodFilled, param.Pods = matchWorkload(request.PathParameter("workload_name"), param.Namespaces)
			param.PodFilled, param.Pods = matchPod(request.PathParameter("pod_name"), param.PodFilled, param.Pods)
			param.ContainerFilled, param.Containers = matchContainer(request.PathParameter("container_name"))
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

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

package log

import (
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"reflect"
	"strings"
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

func in(value interface{}, container interface{}) int {
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

func getWorkloadName(name string, kind string) string {
	if kind == "ReplicaSet" {
		lastIndex := strings.LastIndex(name, "-")
		if lastIndex >= 0 {
			return name[:lastIndex]
		}
	}

	return name
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

func QueryWorkspace(workspaceMatch string, workspaceQuery string) (bool, []string) {
	if workspaceMatch == "" && workspaceQuery == "" {
		return false, nil
	}

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return true, nil
	}

	var namespaces []string

	var hasMatch = false
	var workspacesMatch []string
	if workspaceMatch != "" {
		workspacesMatch = strings.Split(strings.Replace(workspaceMatch, ",", " ", -1), " ")
		hasMatch = true
	}

	var hasQuery = false
	var workspacesQuery []string
	if workspaceQuery != "" {
		workspacesQuery = strings.Split(strings.ToLower(strings.Replace(workspaceQuery, ",", " ", -1)), " ")
		hasQuery = true
	}

	for _, ns := range nsList {
		labels := ns.GetLabels()
		_, ok := labels[constants.WorkspaceLabelKey]
		if ok {
			var namespaceCanAppend = true
			if hasMatch {
				if !matchLabel(labels[constants.WorkspaceLabelKey], workspacesMatch) {
					namespaceCanAppend = false
				}
			}
			if hasQuery {
				if !queryLabel(strings.ToLower(labels[constants.WorkspaceLabelKey]), workspacesQuery) {
					namespaceCanAppend = false
				}
			}

			if namespaceCanAppend {
				namespaces = append(namespaces, ns.GetName())
			}
		}
	}

	return true, namespaces
}

func MatchNamespace(namespaceMatch string, namespaceFilled bool, namespaces []string) (bool, []string) {
	if namespaceMatch == "" {
		return namespaceFilled, namespaces
	}

	namespacesMatch := strings.Split(strings.Replace(namespaceMatch, ",", " ", -1), " ")

	if namespaceFilled {
		return true, intersection(namespacesMatch, namespaces)
	}

	return true, namespacesMatch
}

func QueryWorkload(workloadMatch string, workloadQuery string, namespaces []string) (bool, []string) {
	if workloadMatch == "" && workloadQuery == "" {
		return false, nil
	}

	podLister := informers.SharedInformerFactory().Core().V1().Pods().Lister()
	podList, err := podLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list pods, error: ", err)
		return true, nil
	}

	var pods []string

	var hasMatch = false
	var workloadsMatch []string
	if workloadMatch != "" {
		workloadsMatch = strings.Split(strings.Replace(workloadMatch, ",", " ", -1), " ")
		hasMatch = true
	}

	var hasQuery = false
	var workloadsQuery []string
	if workloadQuery != "" {
		workloadsQuery = strings.Split(strings.ToLower(strings.Replace(workloadQuery, ",", " ", -1)), " ")
		hasQuery = true
	}

	if namespaces == nil {
		for _, pod := range podList {
			/*if len(pod.ObjectMeta.OwnerReferences) > 0 {
				glog.Infof("List Pod %v:%v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			}*/
			if len(pod.ObjectMeta.OwnerReferences) > 0 {
				var podCanAppend = true
				workloadName := getWorkloadName(pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
				if hasMatch {
					if !matchLabel(workloadName, workloadsMatch) {
						podCanAppend = false
					}
				}
				if hasQuery {
					if !queryLabel(strings.ToLower(workloadName), workloadsQuery) {
						podCanAppend = false
					}
				}

				if podCanAppend {
					pods = append(pods, pod.Name)
				}
			}
		}
	} else {
		for _, pod := range podList {
			/*if len(pod.ObjectMeta.OwnerReferences) > 0 {
				glog.Infof("List Pod %v:%v:%v", pod.Name, pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
			}*/
			if len(pod.ObjectMeta.OwnerReferences) > 0 && in(pod.Namespace, namespaces) >= 0 {
				var podCanAppend = true
				workloadName := getWorkloadName(pod.ObjectMeta.OwnerReferences[0].Name, pod.ObjectMeta.OwnerReferences[0].Kind)
				if hasMatch {
					if !matchLabel(workloadName, workloadsMatch) {
						podCanAppend = false
					}
				}
				if hasQuery {
					if !queryLabel(strings.ToLower(workloadName), workloadsQuery) {
						podCanAppend = false
					}
				}

				if podCanAppend {
					pods = append(pods, pod.Name)
				}
			}
		}
	}

	return true, pods
}

func MatchPod(podMatch string, podFilled bool, pods []string) (bool, []string) {
	if podMatch == "" {
		return podFilled, pods
	}

	podsMatch := strings.Split(strings.Replace(podMatch, ",", " ", -1), " ")

	if podFilled {
		return true, intersection(podsMatch, pods)
	}

	return true, podsMatch
}

func MatchContainer(containerMatch string) (bool, []string) {
	if containerMatch == "" {
		return false, nil
	}

	return true, strings.Split(strings.Replace(containerMatch, ",", " ", -1), " ")
}

func GetWorkspaceOfNamesapce(namespace string) string {
	var workspace string
	workspace = ""

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		glog.Error("failed to list namespace, error: ", err)
		return workspace
	}

	for _, ns := range nsList {
		if ns.GetName() == namespace {
			labels := ns.GetLabels()
			_, ok := labels[constants.WorkspaceLabelKey]
			if ok {
				workspace = labels[constants.WorkspaceLabelKey]
			}
		}
	}

	return workspace
}

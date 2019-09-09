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
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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

// list namespaces that match search conditions
func MatchNamespace(nsFilter []string, nsQuery []string, wsFilter []string, wsQuery []string) (bool, []string) {

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		glog.Errorf("failed to list namespace, error: %s", err)
		return true, nil
	}

	var namespaces []string

	// if no search condition is set on both namespace and workspace,
	// then return all namespaces
	if nsQuery == nil && nsFilter == nil && wsQuery == nil && wsFilter == nil {
		for _, ns := range nsList {
			namespaces = append(namespaces, ns.Name)
		}
		return false, namespaces
	}

	for _, ns := range nsList {
		if stringutils.StringIn(ns.Name, nsFilter) ||
			stringutils.StringIn(ns.Annotations[constants.WorkspaceLabelKey], wsFilter) ||
			containsIn(ns.Name, nsQuery) ||
			containsIn(ns.Annotations[constants.WorkspaceLabelKey], wsQuery) {
			namespaces = append(namespaces, ns.Name)
		}
	}

	// if namespaces is equal to nil, indicates no namespace matched
	// it causes the query to return no result
	return namespaces == nil, namespaces
}

func containsIn(str string, subStrs []string) bool {
	for _, sub := range subStrs {
		if strings.Contains(str, sub) {
			return true
		}
	}
	return false
}

func MakeNamespaceCreationTimeMap(namespaces []string) map[string]string {

	namespaceWithCreationTime := make(map[string]string)

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	for _, item := range namespaces {
		ns, err := nsLister.Get(item)
		if err != nil {
			// the ns doesn't exist
			continue
		}
		namespaceWithCreationTime[ns.Name] = strconv.FormatInt(ns.CreationTimestamp.UnixNano()/int64(time.Millisecond), 10)
	}

	return namespaceWithCreationTime
}

func MatchWorkload(workloadMatch string, workloadQuery string, namespaces []string) (bool, []string) {
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

	// if workloads is equal to nil, indicates no workload matched
	// it causes the query to return no result
	return pods == nil, pods
}

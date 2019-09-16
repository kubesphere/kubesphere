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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
	"strconv"
	"strings"
	"time"
)

// list namespaces that match search conditions
func MatchNamespace(nsFilter []string, nsQuery []string, wsFilter []string, wsQuery []string) (bool, []string) {

	nsLister := informers.SharedInformerFactory().Core().V1().Namespaces().Lister()
	nsList, err := nsLister.List(labels.Everything())
	if err != nil {
		klog.Errorf("failed to list namespace, error: %s", err)
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

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
package resources

import (
	"fmt"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sort"
	"strings"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type nodeSearcher struct {
}

func (*nodeSearcher) get(namespace, name string) (interface{}, error) {
	return informers.SharedInformerFactory().Core().V1().Nodes().Lister().Get(name)
}

func getNodeStatus(node *v1.Node) string {
	if node.Spec.Unschedulable {
		return StatusUnschedulable
	}
	for _, condition := range node.Status.Conditions {
		if isUnhealthStatus(condition) {
			return StatusWarning
		}
	}

	return StatusRunning
}

const NodeConfigOK v1.NodeConditionType = "ConfigOK"
const NodeKubeletReady v1.NodeConditionType = "KubeletReady"

var expectedConditions = map[v1.NodeConditionType]v1.ConditionStatus{
	v1.NodeMemoryPressure:     v1.ConditionFalse,
	v1.NodeDiskPressure:       v1.ConditionFalse,
	v1.NodePIDPressure:        v1.ConditionFalse,
	v1.NodeNetworkUnavailable: v1.ConditionFalse,
	NodeConfigOK:              v1.ConditionTrue,
	NodeKubeletReady:          v1.ConditionTrue,
	v1.NodeReady:              v1.ConditionTrue,
}

func isUnhealthStatus(condition v1.NodeCondition) bool {
	expectedStatus := expectedConditions[condition.Type]
	if expectedStatus != "" && condition.Status != expectedStatus {
		return true
	}
	return false
}

// exactly Match
func (*nodeSearcher) match(match map[string]string, item *v1.Node) bool {
	for k, v := range match {
		switch k {
		case Name:
			names := strings.Split(v, "|")
			if !sliceutil.HasString(names, item.Name) {
				return false
			}
		case Role:
			labelKey := fmt.Sprintf("node-role.kubernetes.io/%s", v)
			if _, ok := item.Labels[labelKey]; !ok {
				return false
			}
		case Status:
			if getNodeStatus(item) != v {
				return false
			}
		case Keyword:
			if !strings.Contains(item.Name, v) && !searchFuzzy(item.Labels, "", v) && !searchFuzzy(item.Annotations, "", v) {
				return false
			}
		default:
			// label not exist or value not equal
			if val, ok := item.Labels[k]; !ok || val != v {
				return false
			}
		}
	}
	return true
}

// Fuzzy searchInNamespace
func (*nodeSearcher) fuzzy(fuzzy map[string]string, item *v1.Node) bool {
	for k, v := range fuzzy {
		switch k {
		case Name:
			if !strings.Contains(item.Name, v) && !strings.Contains(item.Annotations[constants.DisplayNameAnnotationKey], v) {
				return false
			}
		case Label:
			if !searchFuzzy(item.Labels, "", v) {
				return false
			}
		case annotation:
			if !searchFuzzy(item.Annotations, "", v) {
				return false
			}
			return false
		case app:
			if !strings.Contains(item.Labels[chart], v) && !strings.Contains(item.Labels[release], v) {
				return false
			}
		default:
			if !searchFuzzy(item.Labels, k, v) {
				return false
			}
		}
	}
	return true
}

func (*nodeSearcher) compare(a, b *v1.Node, orderBy string) bool {
	switch orderBy {
	case CreateTime:
		return a.CreationTimestamp.Time.Before(b.CreationTimestamp.Time)
	case Name:
		fallthrough
	default:
		return strings.Compare(a.Name, b.Name) <= 0
	}
}

func (s *nodeSearcher) search(namespace string, conditions *params.Conditions, orderBy string, reverse bool) ([]interface{}, error) {
	nodes, err := informers.SharedInformerFactory().Core().V1().Nodes().Lister().List(labels.Everything())

	if err != nil {
		return nil, err
	}

	result := make([]*v1.Node, 0)

	if len(conditions.Match) == 0 && len(conditions.Fuzzy) == 0 {
		result = nodes
	} else {
		for _, item := range nodes {
			if s.match(conditions.Match, item) && s.fuzzy(conditions.Fuzzy, item) {
				result = append(result, item)
			}
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if reverse {
			tmp := i
			i = j
			j = tmp
		}
		return s.compare(result[i], result[j], orderBy)
	})

	r := make([]interface{}, 0)
	for _, i := range result {
		r = append(r, i)
	}
	return r, nil
}

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
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"kubesphere.io/kubesphere/pkg/informers"
	"kubesphere.io/kubesphere/pkg/server/params"
	"sort"
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
func (*nodeSearcher) match(kv map[string]string, item *v1.Node) bool {
	for k, v := range kv {
		switch k {
		case Role:
			labelKey := fmt.Sprintf("node-role.kubernetes.io/%s", v)
			if _, ok := item.Labels[labelKey]; !ok {
				return false
			}
		case Status:
			if getNodeStatus(item) != v {
				return false
			}
		default:
			if !match(k, v, item.ObjectMeta) {
				return false
			}
		}
	}
	return true
}

// Fuzzy searchInNamespace
func (*nodeSearcher) fuzzy(kv map[string]string, item *v1.Node) bool {
	for k, v := range kv {
		if !fuzzy(k, v, item.ObjectMeta) {
			return false
		}
	}
	return true
}

func (*nodeSearcher) compare(a, b *v1.Node, orderBy string) bool {
	return compare(a.ObjectMeta, b.ObjectMeta, orderBy)
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

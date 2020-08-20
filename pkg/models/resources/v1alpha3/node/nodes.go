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

package node

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	resourceheper "k8s.io/kubectl/pkg/util/resource"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"sort"
)

// Those annotations were added to node only for display purposes
const (
	nodeCPURequests                                 = "node.kubesphere.io/cpu-requests"
	nodeMemoryRequests                              = "node.kubesphere.io/memory-requests"
	nodeCPULimits                                   = "node.kubesphere.io/cpu-limits"
	nodeMemoryLimits                                = "node.kubesphere.io/memory-limits"
	nodeCPURequestsFraction                         = "node.kubesphere.io/cpu-requests-fraction"
	nodeCPULimitsFraction                           = "node.kubesphere.io/cpu-limits-fraction"
	nodeMemoryRequestsFraction                      = "node.kubesphere.io/memory-requests-fraction"
	nodeMemoryLimitsFraction                        = "node.kubesphere.io/memory-limits-fraction"
	nodeConfigOK               v1.NodeConditionType = "ConfigOK"
	nodeKubeletReady           v1.NodeConditionType = "KubeletReady"
	statusRunning                                   = "running"
	statusWarning                                   = "warning"
	statusUnschedulable                             = "unschedulable"
)

type nodesGetter struct {
	informers informers.SharedInformerFactory
}

func New(informers informers.SharedInformerFactory) v1alpha3.Interface {
	return &nodesGetter{
		informers: informers,
	}
}

func (c nodesGetter) Get(_, name string) (runtime.Object, error) {
	node, err := c.informers.Core().V1().Nodes().Lister().Get(name)
	if err != nil {
		return nil, err
	}

	// ignore the error, skip annotating process if error happened
	pods, _ := c.informers.Core().V1().Pods().Lister().Pods("").List(labels.Everything())
	c.annotateNode(node, pods)

	return node, nil
}

func (c nodesGetter) List(_ string, q *query.Query) (*api.ListResult, error) {
	nodes, err := c.informers.Core().V1().Nodes().Lister().List(q.Selector())
	if err != nil {
		return nil, err
	}

	var filtered []*v1.Node
	for _, object := range nodes {
		selected := true
		for field, value := range q.Filters {
			if !c.filter(object, query.Filter{Field: field, Value: value}) {
				selected = false
				break
			}
		}

		if selected {
			filtered = append(filtered, object)
		}
	}

	// sort by sortBy field
	sort.Slice(filtered, func(i, j int) bool {
		if !q.Ascending {
			return c.compare(filtered[i], filtered[j], q.SortBy)
		}
		return !c.compare(filtered[i], filtered[j], q.SortBy)
	})

	total := len(filtered)
	if q.Pagination == nil {
		q.Pagination = query.NoPagination
	}
	start, end := q.Pagination.GetValidPagination(total)
	selectedNodes := filtered[start:end]

	// ignore the error, skip annotating process if error happened
	pods, _ := c.informers.Core().V1().Pods().Lister().Pods("").List(labels.Everything())
	var nonTerminatedPodsList []*v1.Pod
	for _, pod := range pods {
		if pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
			nonTerminatedPodsList = append(nonTerminatedPodsList, pod)
		}
	}

	var result = make([]interface{}, 0)
	for _, node := range selectedNodes {
		c.annotateNode(node, nonTerminatedPodsList)
		result = append(result, node)
	}

	return &api.ListResult{
		TotalItems: total,
		Items:      result,
	}, nil
}

func (c nodesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNode, ok := left.(*v1.Node)
	if !ok {
		return false
	}

	rightNode, ok := right.(*v1.Node)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftNode.ObjectMeta, rightNode.ObjectMeta, field)
}

func (c nodesGetter) filter(object runtime.Object, filter query.Filter) bool {
	node, ok := object.(*v1.Node)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return getNodeStatus(node) == string(filter.Value)
	}

	return v1alpha3.DefaultObjectMetaFilter(node.ObjectMeta, filter)
}

// annotateNode adds cpu/memory requests usage data to node's annotations
func (c nodesGetter) annotateNode(node *v1.Node, pods []*v1.Pod) {
	if len(pods) == 0 {
		return
	}

	var nodePods []*v1.Pod
	for _, pod := range pods {
		if pod.Spec.NodeName == node.Name {
			nodePods = append(nodePods, pod)
		}
	}

	reqs, limits := c.getPodsTotalRequestAndLimits(nodePods)

	if node.Annotations == nil {
		node.Annotations = make(map[string]string)
	}

	cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
	node.Annotations[nodeCPURequests] = cpuReqs.String()
	node.Annotations[nodeCPULimits] = cpuLimits.String()
	node.Annotations[nodeMemoryRequests] = memoryReqs.String()
	node.Annotations[nodeMemoryLimits] = memoryLimits.String()

	fractionCpuReqs, fractionCpuLimits := float64(0), float64(0)
	allocatable := node.Status.Allocatable
	if allocatable.Cpu().MilliValue() != 0 {
		fractionCpuReqs = float64(cpuReqs.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
		fractionCpuLimits = float64(cpuLimits.MilliValue()) / float64(allocatable.Cpu().MilliValue()) * 100
	}
	fractionMemoryReqs, fractionMemoryLimits := float64(0), float64(0)
	if allocatable.Memory().Value() != 0 {
		fractionMemoryReqs = float64(memoryReqs.Value()) / float64(allocatable.Memory().Value()) * 100
		fractionMemoryLimits = float64(memoryLimits.Value()) / float64(allocatable.Memory().Value()) * 100
	}

	node.Annotations[nodeCPURequestsFraction] = fmt.Sprintf("%d%%", int(fractionCpuReqs))
	node.Annotations[nodeCPULimitsFraction] = fmt.Sprintf("%d%%", int(fractionCpuLimits))
	node.Annotations[nodeMemoryRequestsFraction] = fmt.Sprintf("%d%%", int(fractionMemoryReqs))
	node.Annotations[nodeMemoryLimitsFraction] = fmt.Sprintf("%d%%", int(fractionMemoryLimits))
}

func (c nodesGetter) getPodsTotalRequestAndLimits(pods []*v1.Pod) (reqs map[v1.ResourceName]resource.Quantity, limits map[v1.ResourceName]resource.Quantity) {
	reqs, limits = map[v1.ResourceName]resource.Quantity{}, map[v1.ResourceName]resource.Quantity{}
	for _, pod := range pods {
		podReqs, podLimits := resourceheper.PodRequestsAndLimits(pod)
		for podReqName, podReqValue := range podReqs {
			if value, ok := reqs[podReqName]; !ok {
				reqs[podReqName] = podReqValue.DeepCopy()
			} else {
				value.Add(podReqValue)
				reqs[podReqName] = value
			}
		}
		for podLimitName, podLimitValue := range podLimits {
			if value, ok := limits[podLimitName]; !ok {
				limits[podLimitName] = podLimitValue.DeepCopy()
			} else {
				value.Add(podLimitValue)
				limits[podLimitName] = value
			}
		}
	}
	return
}

func getNodeStatus(node *v1.Node) string {
	if node.Spec.Unschedulable {
		return statusUnschedulable
	}
	for _, condition := range node.Status.Conditions {
		if isUnhealthyStatus(condition) {
			return statusWarning
		}
	}

	return statusRunning
}

var expectedConditions = map[v1.NodeConditionType]v1.ConditionStatus{
	v1.NodeMemoryPressure:     v1.ConditionFalse,
	v1.NodeDiskPressure:       v1.ConditionFalse,
	v1.NodePIDPressure:        v1.ConditionFalse,
	v1.NodeNetworkUnavailable: v1.ConditionFalse,
	nodeConfigOK:              v1.ConditionTrue,
	nodeKubeletReady:          v1.ConditionTrue,
	v1.NodeReady:              v1.ConditionTrue,
}

func isUnhealthyStatus(condition v1.NodeCondition) bool {
	expectedStatus := expectedConditions[condition.Type]
	if expectedStatus != "" && condition.Status != expectedStatus {
		return true
	}
	return false
}

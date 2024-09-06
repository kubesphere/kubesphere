/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package node

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

// Those annotations were added to node only for display purposes
const (
	nodeCPURequests                                     = "node.kubesphere.io/cpu-requests"
	nodeMemoryRequests                                  = "node.kubesphere.io/memory-requests"
	nodeCPULimits                                       = "node.kubesphere.io/cpu-limits"
	nodeMemoryLimits                                    = "node.kubesphere.io/memory-limits"
	nodeCPURequestsFraction                             = "node.kubesphere.io/cpu-requests-fraction"
	nodeCPULimitsFraction                               = "node.kubesphere.io/cpu-limits-fraction"
	nodeMemoryRequestsFraction                          = "node.kubesphere.io/memory-requests-fraction"
	nodeMemoryLimitsFraction                            = "node.kubesphere.io/memory-limits-fraction"
	nodeConfigOK               corev1.NodeConditionType = "ConfigOK"
	nodeKubeletReady           corev1.NodeConditionType = "KubeletReady"
	statusRunning                                       = "running"
	statusWarning                                       = "warning"
	statusUnschedulable                                 = "unschedulable"
)

type nodesGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &nodesGetter{cache: cache}
}

func (c *nodesGetter) Get(_, name string) (runtime.Object, error) {
	node := &corev1.Node{}
	return node, c.cache.Get(context.Background(), types.NamespacedName{Name: name}, node)
}

func (c *nodesGetter) List(_ string, q *query.Query) (*api.ListResult, error) {
	nodes := &corev1.NodeList{}
	if err := c.cache.List(context.Background(), nodes,
		client.MatchingLabelsSelector{Selector: q.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range nodes.Items {
		result = append(result, item.DeepCopy())
	}
	return v1alpha3.DefaultList(result, q, c.compare, c.filter), nil
}

func (c *nodesGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftNode, ok := left.(*corev1.Node)
	if !ok {
		return false
	}

	rightNode, ok := right.(*corev1.Node)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftNode.ObjectMeta, rightNode.ObjectMeta, field)
}

func (c *nodesGetter) filter(object runtime.Object, filter query.Filter) bool {
	node, ok := object.(*corev1.Node)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return getNodeStatus(node) == string(filter.Value)
	}
	return v1alpha3.DefaultObjectMetaFilter(node.ObjectMeta, filter)
}

func getNodeStatus(node *corev1.Node) string {
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

var expectedConditions = map[corev1.NodeConditionType]corev1.ConditionStatus{
	corev1.NodeMemoryPressure:     corev1.ConditionFalse,
	corev1.NodeDiskPressure:       corev1.ConditionFalse,
	corev1.NodePIDPressure:        corev1.ConditionFalse,
	corev1.NodeNetworkUnavailable: corev1.ConditionFalse,
	nodeConfigOK:                  corev1.ConditionTrue,
	nodeKubeletReady:              corev1.ConditionTrue,
	corev1.NodeReady:              corev1.ConditionTrue,
}

func isUnhealthyStatus(condition corev1.NodeCondition) bool {
	expectedStatus := expectedConditions[condition.Type]
	if expectedStatus != "" && condition.Status != expectedStatus {
		return true
	}
	return false
}

package node

import (
    v1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/apimachinery/pkg/labels"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/informers"
    resourceheper "k8s.io/kubectl/pkg/util/resource"
    "kubesphere.io/kubesphere/pkg/api"
    clusterv1alpha1 "kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
    "kubesphere.io/kubesphere/pkg/apiserver/query"
    "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

// Those annotations were added to node only for display purposes
const (
    nodeCPURequests    = "node.kubesphere.io/cpu-requests"
    nodeMemoryRequests = "node.kubesphere.io/memory-requests"
    nodeCPULimits    = "node.kubesphere.io/cpu-limits"
    nodeMemoryLimits = "node.kubesphere.io/memory-limits"
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
    return c.informers.Core().V1().Nodes().Lister().Get(name)
}

func (c nodesGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
    nodes, err := c.informers.Core().V1().Nodes().Lister().List(query.Selector())
    if err != nil {
        return nil, err
    }

    // ignore the error, skip annotating process if error happened
    pods, _ := c.informers.Core().V1().Pods().Lister().Pods("").List(labels.Everything())

    var result []runtime.Object
    for _, node := range nodes {
        c.annotateNode(node, pods)
        result = append(result, node)
    }

    return v1alpha3.DefaultList(result, query, c.compare, c.filter), nil
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
    cluster, ok := object.(*clusterv1alpha1.Cluster)
    if !ok {
        return false
    }

    return v1alpha3.DefaultObjectMetaFilter(cluster.ObjectMeta, filter)
}

func (c nodesGetter) annotateNode(node *v1.Node, pods []*v1.Pod) {
    if len(pods) == 0 {
        return
    }

    var nodeNonTerminatedPodsList []*v1.Pod
    for _, pod := range pods {
        if pod.Spec.NodeName == node.Name && pod.Status.Phase != v1.PodSucceeded && pod.Status.Phase != v1.PodFailed {
            nodeNonTerminatedPodsList = append(nodeNonTerminatedPodsList, pod)
        }
    }

    reqs, limits := c.getPodsTotalRequestAndLimits(nodeNonTerminatedPodsList)

    if node.Annotations == nil {
        node.Annotations = make(map[string]string)
    }

    cpuReqs, cpuLimits, memoryReqs, memoryLimits := reqs[v1.ResourceCPU], limits[v1.ResourceCPU], reqs[v1.ResourceMemory], limits[v1.ResourceMemory]
    node.Annotations[nodeCPURequests] = cpuReqs.String()
    node.Annotations[nodeCPULimits] = cpuLimits.String()
    node.Annotations[nodeMemoryRequests] = memoryReqs.String()
    node.Annotations[nodeMemoryLimits] = memoryLimits.String()
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
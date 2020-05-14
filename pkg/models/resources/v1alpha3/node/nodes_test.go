package node

import (
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

// mergeResourceLists will merge resoure lists. When two lists have the same resourece, the value from
// the last list will be present in the result
func mergeResourceLists(resourceLists ...corev1.ResourceList) corev1.ResourceList {
	result := corev1.ResourceList{}
	for _, rl := range resourceLists {
		for resource, quantity := range rl {
			result[resource] = quantity
		}
	}
	return result
}

func getResourceList(cpu, memory string) corev1.ResourceList {
	res := corev1.ResourceList{}
	if cpu != "" {
		res[corev1.ResourceCPU] = resource.MustParse(cpu)
	}
	if memory != "" {
		res[corev1.ResourceMemory] = resource.MustParse(memory)
	}
	return res
}

var nodeAllocatable = mergeResourceLists(getResourceList("4", "12Gi"))
var node = &corev1.Node{
	ObjectMeta: metav1.ObjectMeta{
		Name: "foo",
	},
	Status: corev1.NodeStatus{
		Allocatable: nodeAllocatable,
	},
}

var pods = []*corev1.Pod{
	{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo",
			Name:      "pod-with-resources",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: corev1.PodSpec{
			NodeName: node.Name,
			Containers: []corev1.Container{
				{
					Name:  "cpu-mem",
					Image: "image:latest",
					Resources: corev1.ResourceRequirements{
						Requests: getResourceList("1", "1Gi"),
						Limits:   getResourceList("2", "2Gi"),
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	},
	{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "foo2",
			Name:      "pod-with-resources",
		},
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
		},
		Spec: corev1.PodSpec{
			NodeName: node.Name,
			Containers: []corev1.Container{
				{
					Name:  "cpu-mem",
					Image: "image:latest",
					Resources: corev1.ResourceRequirements{
						Requests: getResourceList("1", "1Gi"),
						Limits:   getResourceList("2", "2Gi"),
					},
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	},
}

var expectedAnnotations = map[string]string{
	nodeCPURequests:            "2",
	nodeCPULimits:              "4",
	nodeCPURequestsFraction:    "50%",
	nodeCPULimitsFraction:      "100%",
	nodeMemoryRequests:         "2Gi",
	nodeMemoryLimits:           "4Gi",
	nodeMemoryRequestsFraction: "16%",
	nodeMemoryLimitsFraction:   "33%",
}

func TestNodesGetterGet(t *testing.T) {
	fake := fake.NewSimpleClientset(node, pods[0], pods[1])

	informer := informers.NewSharedInformerFactory(fake, 0)
	informer.Core().V1().Nodes().Informer().GetIndexer().Add(node)
	for _, pod := range pods {
		informer.Core().V1().Pods().Informer().GetIndexer().Add(pod)
	}

	nodeGetter := New(informer)
	got, err := nodeGetter.Get("", node.Name)
	if err != nil {
		t.Fatal(err)
	}
	nodeGot := got.(*corev1.Node)

	if diff := cmp.Diff(nodeGot.Annotations, expectedAnnotations); len(diff) != 0 {
		t.Errorf("%T, diff(-got, +expected), %v", expectedAnnotations, nodeGot.Annotations)
	}

}

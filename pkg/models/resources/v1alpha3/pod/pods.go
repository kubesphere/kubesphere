/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package pod

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

const (
	fieldNodeName    = "nodeName"
	fieldPVCName     = "pvcName"
	fieldServiceName = "serviceName"
	fieldPhase       = "phase"
	fieldStatus      = "status"
	fieldPodIP       = "podIP"

	statusTypeWaitting  = "Waiting"
	statusTypeRunning   = "Running"
	statusTypeError     = "Error"
	statusTypeCompleted = "Completed"
)

type podsGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &podsGetter{cache: cache}
}

func (p *podsGetter) Get(namespace, name string) (runtime.Object, error) {
	pod := &corev1.Pod{}
	if err := p.cache.Get(context.Background(), types.NamespacedName{Namespace: namespace, Name: name}, pod); err != nil {
		return nil, err
	}
	return p.setPodStatus(pod.DeepCopy()), nil
}

func (p *podsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	pods := &corev1.PodList{}
	if err := p.cache.List(context.Background(), pods, client.InNamespace(namespace),
		client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
		return nil, err
	}
	var result []runtime.Object
	for _, item := range pods.Items {
		result = append(result, p.setPodStatus(item.DeepCopy()))
	}
	return v1alpha3.DefaultList(result, query, p.compare, p.filter), nil
}

func (p *podsGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftPod, ok := left.(*corev1.Pod)
	if !ok {
		return false
	}

	rightPod, ok := right.(*corev1.Pod)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftPod.ObjectMeta, rightPod.ObjectMeta, field)
}

func (p *podsGetter) filter(object runtime.Object, filter query.Filter) bool {
	pod, ok := object.(*corev1.Pod)

	if !ok {
		return false
	}
	switch filter.Field {
	case fieldNodeName:
		return pod.Spec.NodeName == string(filter.Value)
	case fieldPVCName:
		return p.podBindPVC(pod, string(filter.Value))
	case fieldServiceName:
		return p.podBelongToService(pod, string(filter.Value))
	case fieldStatus:
		return p.getPodStatus(pod) == string(filter.Value)
	case fieldPhase:
		return string(pod.Status.Phase) == string(filter.Value)
	case fieldPodIP:
		return p.podWithIP(pod, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
}

func (p *podsGetter) podWithIP(item *corev1.Pod, ipAddress string) bool {
	for _, ip := range item.Status.PodIPs {
		if strings.Contains(ip.String(), ipAddress) {
			return true
		}
	}
	return false
}

func (p *podsGetter) podBindPVC(item *corev1.Pod, pvcName string) bool {
	for _, v := range item.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName == pvcName {
			return true
		}
	}
	return false
}

func (p *podsGetter) podBelongToService(item *corev1.Pod, serviceName string) bool {
	service := &corev1.Service{}
	if err := p.cache.Get(context.Background(), types.NamespacedName{Namespace: item.Namespace, Name: serviceName}, service); err != nil {
		return false
	}
	selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
	if selector.Empty() || !selector.Matches(labels.Set(item.Labels)) {
		return false
	}
	return true
}

func (p *podsGetter) setPodStatus(pod *corev1.Pod) *corev1.Pod {
	pod.Status.Phase = corev1.PodPhase(p.getPodStatus(pod))
	return pod
}

// getPodStatus refer to `kubectl get po` result.
// https://github.com/kubernetes/kubernetes/blob/45279654db87f4908911569c07afc42804f0e246/pkg/printers/internalversion/printers.go#L820-920
// podStatusPhase 			  = []string("Pending", "Running","Succeeded","Failed","Unknown")
// podStatusReasons           = []string{"Evicted", "NodeAffinity", "NodeLost", "Shutdown", "UnexpectedAdmissionError"}
// containerWaitingReasons    = []string{"ContainerCreating", "CrashLoopBackOff", "CreateContainerConfigError", "ErrImagePull", "ImagePullBackOff", "CreateContainerError", "InvalidImageName"}
// containerTerminatedReasons = []string{"OOMKilled", "Completed", "Error", "ContainerCannotRun", "DeadlineExceeded", "Evicted"}
func (p *podsGetter) getPodStatus(pod *corev1.Pod) string {
	reason := string(pod.Status.Phase)

	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	/*
		todo: upgrade k8s.io/api version

		// If the Pod carries {type:PodScheduled, reason:WaitingForGates}, set reason to 'SchedulingGated'.
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodScheduled && condition.Reason == corev1.PodReasonSchedulingGated {
				reason = corev1.PodReasonSchedulingGated
			}
		}
	*/

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]

		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {

		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true

			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			if hasPodReadyCondition(pod.Status.Conditions) {
				reason = "Running"
			} else {
				reason = "NotReady"
			}
		}
	}

	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}

	statusType := statusTypeWaitting
	switch reason {
	case "Running":
		statusType = statusTypeRunning
	case "Failed":
		statusType = statusTypeError
	case "Error":
		statusType = statusTypeError
	case "Completed":
		statusType = statusTypeCompleted
	case "Succeeded":
		if isPodReadyConditionReason(pod.Status.Conditions, "PodCompleted") {
			statusType = statusTypeCompleted
		}
	default:
		if strings.HasPrefix(reason, "OutOf") {
			statusType = statusTypeError
		}
	}
	return statusType
}

func hasPodReadyCondition(conditions []corev1.PodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isPodReadyConditionReason(conditions []corev1.PodCondition, reason string) bool {
	for _, condition := range conditions {
		if condition.Type == corev1.PodReady && condition.Reason != reason {
			return false
		}
	}
	return true
}

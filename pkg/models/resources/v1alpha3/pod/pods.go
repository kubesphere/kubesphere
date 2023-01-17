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

package pod

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"

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

	statusTypeWaitting  = "Waiting"
	statusTypeRunning   = "Running"
	statusTypeError     = "Error"
	statusTypeCompleted = "Completed"
)

type podsGetter struct {
	informer informers.SharedInformerFactory
}

func New(sharedInformers informers.SharedInformerFactory) v1alpha3.Interface {
	return &podsGetter{informer: sharedInformers}
}

func (p *podsGetter) Get(namespace, name string) (runtime.Object, error) {
	return p.informer.Core().V1().Pods().Lister().Pods(namespace).Get(name)
}

func (p *podsGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {

	pods, err := p.informer.Core().V1().Pods().Lister().Pods(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, pod := range pods {
		result = append(result, pod)
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
		_, statusType := p.getPodStatus(pod)
		return statusType == string(filter.Value)
	case fieldPhase:
		return string(pod.Status.Phase) == string(filter.Value)
	default:
		return v1alpha3.DefaultObjectMetaFilter(pod.ObjectMeta, filter)
	}
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
	service, err := p.informer.Core().V1().Services().Lister().Services(item.Namespace).Get(serviceName)
	if err != nil {
		return false
	}
	selector := labels.Set(service.Spec.Selector).AsSelectorPreValidated()
	if selector.Empty() || !selector.Matches(labels.Set(item.Labels)) {
		return false
	}
	return true
}

// getPodStatus refer to `kubectl get po` result.
// https://github.com/kubernetes/kubernetes/blob/45279654db87f4908911569c07afc42804f0e246/pkg/printers/internalversion/printers.go#L820-920
// podStatusPhase 			  = []string("Pending", "Running","Succeeded","Failed","Unknown")
// podStatusReasons           = []string{"Evicted", "NodeAffinity", "NodeLost", "Shutdown", "UnexpectedAdmissionError"}
// containerWaitingReasons    = []string{"ContainerCreating", "CrashLoopBackOff", "CreateContainerConfigError", "ErrImagePull", "ImagePullBackOff", "CreateContainerError", "InvalidImageName"}
// containerTerminatedReasons = []string{"OOMKilled", "Completed", "Error", "ContainerCannotRun", "DeadlineExceeded", "Evicted"}
func (p *podsGetter) getPodStatus(pod *corev1.Pod) (string, string) {
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
	return reason, statusType
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

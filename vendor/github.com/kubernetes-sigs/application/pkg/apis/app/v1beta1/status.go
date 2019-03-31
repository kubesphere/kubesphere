/*
Copyright 2018 The Kubernetes Authors
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

package v1beta1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Constants defining labels
const (
	StatusReady      = "Ready"
	StatusInProgress = "InProgress"
	StatusDisabled   = "Disabled"
)

func (s *ObjectStatus) update(rsrc metav1.Object) {
	ro := rsrc.(runtime.Object)
	gvk := ro.GetObjectKind().GroupVersionKind()
	s.Link = rsrc.GetSelfLink()
	s.Name = rsrc.GetName()
	s.Group = gvk.GroupVersion().String()
	s.Kind = gvk.GroupKind().Kind
	s.Status = StatusReady
}

// ResetComponentList - reset component list objects
func (m *ApplicationStatus) ResetComponentList() {
	m.ComponentList.Objects = []ObjectStatus{}
}

// UpdateStatus the component status
func (m *ApplicationStatus) UpdateStatus(rsrcs []metav1.Object, err error) {
	var ready = true
	for _, r := range rsrcs {
		os := ObjectStatus{}
		os.update(r)
		switch r.(type) {
		case *appsv1.StatefulSet:
			os.Status = stsStatus(r.(*appsv1.StatefulSet))
		case *policyv1.PodDisruptionBudget:
			os.Status = pdbStatus(r.(*policyv1.PodDisruptionBudget))
		case *appsv1.Deployment:
			os.Status = deploymentStatus(r.(*appsv1.Deployment))
		case *appsv1.ReplicaSet:
			os.Status = replicasetStatus(r.(*appsv1.ReplicaSet))
		case *appsv1.DaemonSet:
			os.Status = daemonsetStatus(r.(*appsv1.DaemonSet))
		case *corev1.Pod:
			os.Status = podStatus(r.(*corev1.Pod))
		case *corev1.Service:
			os.Status = serviceStatus(r.(*corev1.Service))
		case *corev1.PersistentVolumeClaim:
			os.Status = pvcStatus(r.(*corev1.PersistentVolumeClaim))
		//case *corev1.ReplicationController:
		// Ingress
		default:
			os.Status = StatusReady
		}
		m.ComponentList.Objects = append(m.ComponentList.Objects, os)
	}
	for _, os := range m.ComponentList.Objects {
		if os.Status != StatusReady {
			ready = false
		}
	}

	if ready {
		m.Ready("ComponentsReady", "all components ready")
	} else {
		m.NotReady("ComponentsNotReady", "some components not ready")
	}
	if err != nil {
		m.SetCondition(Error, "ErrorSeen", err.Error())
	}
}

// Resource specific logic -----------------------------------

// Statefulset
func stsStatus(rsrc *appsv1.StatefulSet) string {
	if rsrc.Status.ReadyReplicas == *rsrc.Spec.Replicas && rsrc.Status.CurrentReplicas == *rsrc.Spec.Replicas {
		return StatusReady
	}
	return StatusInProgress
}

// Deployment
func deploymentStatus(rsrc *appsv1.Deployment) string {
	status := StatusInProgress
	progress := true
	available := true
	for _, c := range rsrc.Status.Conditions {
		switch c.Type {
		case appsv1.DeploymentProgressing:
			// https://github.com/kubernetes/kubernetes/blob/a3ccea9d8743f2ff82e41b6c2af6dc2c41dc7b10/pkg/controller/deployment/progress.go#L52
			if c.Status != corev1.ConditionTrue || c.Reason != "NewReplicaSetAvailable" {
				progress = false
			}
		case appsv1.DeploymentAvailable:
			if c.Status == corev1.ConditionFalse {
				available = false
			}
		}
	}

	if progress && available {
		status = StatusReady
	}

	return status
}

// Replicaset
func replicasetStatus(rsrc *appsv1.ReplicaSet) string {
	status := StatusInProgress
	failure := false
	for _, c := range rsrc.Status.Conditions {
		switch c.Type {
		// https://github.com/kubernetes/kubernetes/blob/a3ccea9d8743f2ff82e41b6c2af6dc2c41dc7b10/pkg/controller/replicaset/replica_set_utils.go
		case appsv1.ReplicaSetReplicaFailure:
			if c.Status == corev1.ConditionTrue {
				failure = true
				break
			}
		}
	}

	if !failure && rsrc.Status.ReadyReplicas == rsrc.Status.Replicas && rsrc.Status.Replicas == rsrc.Status.AvailableReplicas {
		status = StatusReady
	}

	return status
}

// Daemonset
func daemonsetStatus(rsrc *appsv1.DaemonSet) string {
	status := StatusInProgress
	if rsrc.Status.DesiredNumberScheduled == rsrc.Status.NumberAvailable && rsrc.Status.DesiredNumberScheduled == rsrc.Status.NumberReady {
		status = StatusReady
	}
	return status
}

// PVC
func pvcStatus(rsrc *corev1.PersistentVolumeClaim) string {
	status := StatusInProgress
	if rsrc.Status.Phase == corev1.ClaimBound {
		status = StatusReady
	}
	return status
}

// Service
func serviceStatus(rsrc *corev1.Service) string {
	status := StatusReady
	return status
}

// Pod
func podStatus(rsrc *corev1.Pod) string {
	status := StatusInProgress
	for i := range rsrc.Status.Conditions {
		if rsrc.Status.Conditions[i].Type == corev1.PodReady &&
			rsrc.Status.Conditions[i].Status == corev1.ConditionTrue {
			status = StatusReady
			break
		}
	}
	return status
}

// PodDisruptionBudget
func pdbStatus(rsrc *policyv1.PodDisruptionBudget) string {
	if rsrc.Status.CurrentHealthy >= rsrc.Status.DesiredHealthy {
		return StatusReady
	}
	return StatusInProgress
}

// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Constants defining labels
const (
	StatusReady      = "Ready"
	StatusInProgress = "InProgress"
	StatusUnknown    = "Unknown"
	StatusDisabled   = "Disabled"
)

func status(u *unstructured.Unstructured) (string, error) {
	gk := u.GroupVersionKind().GroupKind()
	switch gk.String() {
	case "StatefulSet.apps":
		return stsStatus(u)
	case "Deployment.apps":
		return deploymentStatus(u)
	case "ReplicaSet.apps":
		return replicasetStatus(u)
	case "DaemonSet.apps":
		return daemonsetStatus(u)
	case "PersistentVolumeClaim":
		return pvcStatus(u)
	case "Service":
		return serviceStatus(u)
	case "Pod":
		return podStatus(u)
	case "PodDisruptionBudget.policy":
		return pdbStatus(u)
	case "ReplicationController":
		return replicationControllerStatus(u)
	case "Job.batch":
		return jobStatus(u)
	default:
		return statusFromStandardConditions(u)
	}
}

// Status from standard conditions
func statusFromStandardConditions(u *unstructured.Unstructured) (string, error) {
	condition := StatusReady

	// Check Ready condition
	_, cs, found, err := getConditionOfType(u, StatusReady)
	if err != nil {
		return StatusUnknown, err
	}
	if found && cs == corev1.ConditionFalse {
		condition = StatusInProgress
	}

	// Check InProgress condition
	_, cs, found, err = getConditionOfType(u, StatusInProgress)
	if err != nil {
		return StatusUnknown, err
	}
	if found && cs == corev1.ConditionTrue {
		condition = StatusInProgress
	}

	return condition, nil
}

// Statefulset
func stsStatus(u *unstructured.Unstructured) (string, error) {
	sts := &appsv1.StatefulSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, sts); err != nil {
		return StatusUnknown, err
	}

	if sts.Status.ObservedGeneration == sts.Generation &&
		sts.Status.Replicas == *sts.Spec.Replicas &&
		sts.Status.ReadyReplicas == *sts.Spec.Replicas &&
		sts.Status.CurrentReplicas == *sts.Spec.Replicas {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Deployment
func deploymentStatus(u *unstructured.Unstructured) (string, error) {
	deployment := &appsv1.Deployment{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, deployment); err != nil {
		return StatusUnknown, err
	}

	replicaFailure := false
	progressing := false
	available := false

	for _, condition := range deployment.Status.Conditions {
		switch condition.Type {
		case appsv1.DeploymentProgressing:
			if condition.Status == corev1.ConditionTrue && condition.Reason == "NewReplicaSetAvailable" {
				progressing = true
			}
		case appsv1.DeploymentAvailable:
			if condition.Status == corev1.ConditionTrue {
				available = true
			}
		case appsv1.DeploymentReplicaFailure:
			if condition.Status == corev1.ConditionTrue {
				replicaFailure = true
				break
			}
		}
	}

	if deployment.Status.ObservedGeneration == deployment.Generation &&
		deployment.Status.Replicas == *deployment.Spec.Replicas &&
		deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
		deployment.Status.AvailableReplicas == *deployment.Spec.Replicas &&
		deployment.Status.Conditions != nil && len(deployment.Status.Conditions) > 0 &&
		(progressing || available) && !replicaFailure {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Replicaset
func replicasetStatus(u *unstructured.Unstructured) (string, error) {
	rs := &appsv1.ReplicaSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, rs); err != nil {
		return StatusUnknown, err
	}

	replicaFailure := false
	for _, condition := range rs.Status.Conditions {
		switch condition.Type {
		case appsv1.ReplicaSetReplicaFailure:
			if condition.Status == corev1.ConditionTrue {
				replicaFailure = true
				break
			}
		}
	}
	if rs.Status.ObservedGeneration == rs.Generation &&
		rs.Status.Replicas == *rs.Spec.Replicas &&
		rs.Status.ReadyReplicas == *rs.Spec.Replicas &&
		rs.Status.AvailableReplicas == *rs.Spec.Replicas && !replicaFailure {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Daemonset
func daemonsetStatus(u *unstructured.Unstructured) (string, error) {
	ds := &appsv1.DaemonSet{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, ds); err != nil {
		return StatusUnknown, err
	}

	if ds.Status.ObservedGeneration == ds.Generation &&
		ds.Status.DesiredNumberScheduled == ds.Status.NumberAvailable &&
		ds.Status.DesiredNumberScheduled == ds.Status.NumberReady {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// PVC
func pvcStatus(u *unstructured.Unstructured) (string, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, pvc); err != nil {
		return StatusUnknown, err
	}

	if pvc.Status.Phase == corev1.ClaimBound {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Service
func serviceStatus(u *unstructured.Unstructured) (string, error) {
	service := &corev1.Service{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, service); err != nil {
		return StatusUnknown, err
	}
	stype := service.Spec.Type

	if stype == corev1.ServiceTypeClusterIP || stype == corev1.ServiceTypeNodePort || stype == corev1.ServiceTypeExternalName ||
		stype == corev1.ServiceTypeLoadBalancer && isEmpty(service.Spec.ClusterIP) &&
			len(service.Status.LoadBalancer.Ingress) > 0 && !hasEmptyIngressIP(service.Status.LoadBalancer.Ingress) {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Pod
func podStatus(u *unstructured.Unstructured) (string, error) {
	pod := &corev1.Pod{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, pod); err != nil {
		return StatusUnknown, err
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && (condition.Reason == "PodCompleted" || condition.Status == corev1.ConditionTrue) {
			return StatusReady, nil
		}
	}
	return StatusInProgress, nil
}

// PodDisruptionBudget
func pdbStatus(u *unstructured.Unstructured) (string, error) {
	pdb := &policyv1beta1.PodDisruptionBudget{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, pdb); err != nil {
		return StatusUnknown, err
	}

	if pdb.Status.ObservedGeneration == pdb.Generation &&
		pdb.Status.CurrentHealthy >= pdb.Status.DesiredHealthy {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

func replicationControllerStatus(u *unstructured.Unstructured) (string, error) {
	rc := &corev1.ReplicationController{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, rc); err != nil {
		return StatusUnknown, err
	}

	if rc.Status.ObservedGeneration == rc.Generation &&
		rc.Status.Replicas == *rc.Spec.Replicas &&
		rc.Status.ReadyReplicas == *rc.Spec.Replicas &&
		rc.Status.AvailableReplicas == *rc.Spec.Replicas {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

func jobStatus(u *unstructured.Unstructured) (string, error) {
	job := &batchv1.Job{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, job); err != nil {
		return StatusUnknown, err
	}

	if job.Status.StartTime == nil {
		return StatusInProgress, nil
	}

	return StatusReady, nil
}

func hasEmptyIngressIP(ingress []corev1.LoadBalancerIngress) bool {
	for _, i := range ingress {
		if isEmpty(i.IP) {
			return true
		}
	}
	return false
}

func isEmpty(s string) bool {
	return len(strings.TrimSpace(s)) == 0
}

func getConditionOfType(u *unstructured.Unstructured, conditionType string) (string, corev1.ConditionStatus, bool, error) {
	conditions, found, err := unstructured.NestedSlice(u.Object, "status", "conditions")
	if err != nil || !found {
		return "", corev1.ConditionFalse, false, err
	}

	for _, c := range conditions {
		condition, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		t, found := condition["type"]
		if !found {
			continue
		}
		condType, ok := t.(string)
		if !ok {
			continue
		}
		if condType == conditionType {
			reason := condition["reason"].(string)
			conditionStatus := condition["status"].(string)
			return reason, corev1.ConditionStatus(conditionStatus), true, nil
		}
	}
	return "", corev1.ConditionFalse, false, nil
}

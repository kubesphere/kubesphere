// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Constants defining labels
const (
	StatusReady      = "Ready"
	StatusInProgress = "InProgress"
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
		return "", err
	}
	if found && cs == corev1.ConditionFalse {
		condition = StatusInProgress
	}

	// Check InProgress condition
	_, cs, found, err = getConditionOfType(u, StatusInProgress)
	if err != nil {
		return "", err
	}
	if found && cs == corev1.ConditionTrue {
		condition = StatusInProgress
	}

	return condition, nil
}

// Statefulset
func stsStatus(sts *unstructured.Unstructured) (string, error) {
	readyReplicas, err := getNestedInt64(sts, "status", "readyReplicas")
	if err != nil {
		return "", err
	}

	currentReplicas, err := getNestedInt64(sts, "status", "currentReplicas")
	if err != nil {
		return "", err
	}

	replicas, found, err := unstructured.NestedInt64(sts.Object, "spec", "replicas")
	if err != nil {
		return "", err
	}
	if !found {
		replicas = 1 // This is the default value in the controller if the field is not set.
	}

	if readyReplicas == replicas && currentReplicas == replicas {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Deployment
func deploymentStatus(u *unstructured.Unstructured) (string, error) {
	status := StatusInProgress
	progress := true
	available := true

	reason, conditionStatus, found, err := getConditionOfType(u, string(appsv1.DeploymentProgressing))
	if err != nil {
		return "", err
	}
	if found {
		if conditionStatus != corev1.ConditionTrue || reason != "NewReplicaSetAvailable" {
			progress = false
		}
	}

	_, conditionStatus, found, err = getConditionOfType(u, string(appsv1.DeploymentAvailable))
	if err != nil {
		return "", err
	}
	if found {
		if conditionStatus == corev1.ConditionFalse {
			available = false
		}
	}

	if progress && available {
		status = StatusReady
	}

	return status, nil
}

// Replicaset
func replicasetStatus(u *unstructured.Unstructured) (string, error) {
	status := StatusInProgress

	_, conditionStatus, found, err := getConditionOfType(u, string(appsv1.ReplicaSetReplicaFailure))
	if err != nil {
		return "", err
	}
	if found {
		if conditionStatus == corev1.ConditionTrue {
			return status, nil
		}
	}

	readyReplicas, err := getNestedInt64(u, "status", "readyReplicas")
	if err != nil {
		return "", err
	}

	availableReplicas, err := getNestedInt64(u, "status", "availableReplicas")
	if err != nil {
		return "", err
	}

	replicas, found, err := unstructured.NestedInt64(u.Object, "spec", "replicas")
	if err != nil {
		return "", err
	}
	if !found {
		replicas = 1 // This is the default value in the controller if the field is not set.
	}

	if readyReplicas == replicas && replicas == availableReplicas {
		status = StatusReady
	}

	return status, nil
}

// Daemonset
func daemonsetStatus(u *unstructured.Unstructured) (string, error) {
	desiredNumberScheduled, err := getNestedInt64(u, "status", "desiredNumberScheduled")
	if err != nil {
		return "", err
	}

	numberAvailable, err := getNestedInt64(u, "status", "numberAvailable")
	if err != nil {
		return "", err
	}

	numberReady, err := getNestedInt64(u, "status", "numberReady")
	if err != nil {
		return "", err
	}

	if desiredNumberScheduled == numberAvailable && desiredNumberScheduled == numberReady {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// PVC
func pvcStatus(u *unstructured.Unstructured) (string, error) {
	phase, err := getNestedString(u, "status", "phase")
	if err != nil {
		return "", err
	}

	if phase == string(corev1.ClaimBound) {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

// Service
func serviceStatus(_ *unstructured.Unstructured) (string, error) {
	return StatusReady, nil
}

// Pod
func podStatus(u *unstructured.Unstructured) (string, error) {
	_, conditionStatus, found, err := getConditionOfType(u, string(corev1.PodReady))
	if err != nil {
		return "", err
	}
	if found {
		if conditionStatus == corev1.ConditionTrue {
			return StatusReady, nil
		}
	}
	return StatusInProgress, nil
}

// PodDisruptionBudget
func pdbStatus(u *unstructured.Unstructured) (string, error) {
	currentHealthy, err := getNestedInt64(u, "status", "currentHealthy")
	if err != nil {
		return "", err
	}

	desiredHealthy, err := getNestedInt64(u, "status", "desiredHealthy")
	if err != nil {
		return "", err
	}

	if currentHealthy >= desiredHealthy {
		return StatusReady, nil
	}
	return StatusInProgress, nil
}

func getNestedInt64(u *unstructured.Unstructured, fields ...string) (int64, error) {
	integer, found, err := unstructured.NestedInt64(u.Object, fields...)
	if !found {
		return integer, fmt.Errorf("field %s not found", strings.Join(fields, "."))
	}
	return integer, err
}

func getNestedString(u *unstructured.Unstructured, fields ...string) (string, error) {
	s, found, err := unstructured.NestedString(u.Object, fields...)
	if !found {
		return s, fmt.Errorf("field %s not found", strings.Join(fields, "."))
	}
	return s, err
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

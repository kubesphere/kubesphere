// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appv1beta1 "sigs.k8s.io/application/pkg/apis/app/v1beta1"
)

func setReadyCondition(appStatus *appv1beta1.ApplicationStatus, reason, message string) {
	setCondition(appStatus, appv1beta1.Ready, reason, message)
}

// NotReady - shortcut to set ready condition to false
func setNotReadyCondition(appStatus *appv1beta1.ApplicationStatus, reason, message string) {
	clearCondition(appStatus, appv1beta1.Ready, reason, message)
}

// setErrorCondition - shortcut to set error condition
func setErrorCondition(appStatus *appv1beta1.ApplicationStatus, reason, message string) {
	setCondition(appStatus, appv1beta1.Error, reason, message)
}

// clearErrorCondition - shortcut to set error condition
func clearErrorCondition(appStatus *appv1beta1.ApplicationStatus) {
	clearCondition(appStatus, appv1beta1.Error, "NoError", "No error seen")
}

func setCondition(appStatus *appv1beta1.ApplicationStatus, ctype appv1beta1.ConditionType, reason, message string) {
	setConditionValue(appStatus, ctype, corev1.ConditionTrue, reason, message)
}

func clearCondition(appStatus *appv1beta1.ApplicationStatus, ctype appv1beta1.ConditionType, reason, message string) {
	setConditionValue(appStatus, ctype, corev1.ConditionFalse, reason, message)
}

func setConditionValue(appStatus *appv1beta1.ApplicationStatus, ctype appv1beta1.ConditionType, status corev1.ConditionStatus, reason, message string) {
	var c *appv1beta1.Condition
	for i := range appStatus.Conditions {
		if appStatus.Conditions[i].Type == ctype {
			c = &appStatus.Conditions[i]
		}
	}
	if c == nil {
		addCondition(appStatus, ctype, status, reason, message)
	} else {
		// check message ?
		if c.Status == status && c.Reason == reason && c.Message == message {
			return
		}
		now := metav1.Now()
		c.LastUpdateTime = now
		if c.Status != status {
			c.LastTransitionTime = now
		}
		c.Status = status
		c.Reason = reason
		c.Message = message
	}
}

func addCondition(appStatus *appv1beta1.ApplicationStatus, ctype appv1beta1.ConditionType, status corev1.ConditionStatus, reason, message string) {
	now := metav1.Now()
	c := appv1beta1.Condition{
		Type:               ctype,
		LastUpdateTime:     now,
		LastTransitionTime: now,
		Status:             status,
		Reason:             reason,
		Message:            message,
	}
	appStatus.Conditions = append(appStatus.Conditions, c)
}

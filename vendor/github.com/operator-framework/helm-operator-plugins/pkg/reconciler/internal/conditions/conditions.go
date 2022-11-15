/*
Copyright 2020 The Operator-SDK Authors.

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

package conditions

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"

	"github.com/operator-framework/helm-operator-plugins/pkg/internal/status"
)

const (
	TypeInitialized    = "Initialized"
	TypeDeployed       = "Deployed"
	TypeReleaseFailed  = "ReleaseFailed"
	TypeIrreconcilable = "Irreconcilable"

	ReasonInstallSuccessful   = status.ConditionReason("InstallSuccessful")
	ReasonUpgradeSuccessful   = status.ConditionReason("UpgradeSuccessful")
	ReasonUninstallSuccessful = status.ConditionReason("UninstallSuccessful")

	ReasonErrorGettingClient       = status.ConditionReason("ErrorGettingClient")
	ReasonErrorGettingValues       = status.ConditionReason("ErrorGettingValues")
	ReasonErrorGettingReleaseState = status.ConditionReason("ErrorGettingReleaseState")
	ReasonInstallError             = status.ConditionReason("InstallError")
	ReasonUpgradeError             = status.ConditionReason("UpgradeError")
	ReasonReconcileError           = status.ConditionReason("ReconcileError")
	ReasonUninstallError           = status.ConditionReason("UninstallError")
)

func Initialized(stat corev1.ConditionStatus, reason status.ConditionReason, message interface{}) status.Condition {
	return newCondition(TypeInitialized, stat, reason, message)
}

func Deployed(stat corev1.ConditionStatus, reason status.ConditionReason, message interface{}) status.Condition {
	return newCondition(TypeDeployed, stat, reason, message)
}

func ReleaseFailed(stat corev1.ConditionStatus, reason status.ConditionReason, message interface{}) status.Condition {
	return newCondition(TypeReleaseFailed, stat, reason, message)
}

func Irreconcilable(stat corev1.ConditionStatus, reason status.ConditionReason, message interface{}) status.Condition {
	return newCondition(TypeIrreconcilable, stat, reason, message)
}

func newCondition(t status.ConditionType, s corev1.ConditionStatus, r status.ConditionReason, m interface{}) status.Condition {
	message := fmt.Sprintf("%s", m)
	return status.Condition{
		Type:    t,
		Status:  s,
		Reason:  r,
		Message: message,
	}
}

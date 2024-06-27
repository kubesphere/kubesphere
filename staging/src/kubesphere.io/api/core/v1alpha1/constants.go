/*
Copyright 2022 The KubeSphere Authors.

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

package v1alpha1

const (
	StateUnknown         = ""
	StateInstalling      = "Installing"
	StateInstallFailed   = "InstallFailed"
	StateUpgrading       = "Upgrading"
	StateUpgradeFailed   = "UpgradeFailed"
	StateInstalled       = "Installed"
	StateDeployed        = StateInstalled
	StateUpgraded        = "Upgraded"
	StateUninstalling    = "Uninstalling"
	StateUninstalled     = "Uninstalled"
	StateUninstallFailed = "UninstallFailed"
	// StatePreparing indicates that the Extension is in the Preparing state.
	// This value is only used for Extension objects and is triggered when the state of its InstallPlan is empty
	// and is changing to the Installing/Upgrading state.
	StatePreparing = "Preparing"

	MaxStateNum = 10

	ConditionTypeInitialized = "Initialized"
	ConditionTypeInstalled   = "Installed"
	ConditionTypeUpgraded    = "Upgraded"
	ConditionTypeUninstalled = "Uninstalled"
	ConditionTypeReady       = "Ready"

	DisplayNameAnnotation          = "kubesphere.io/display-name"
	KSVersionAnnotation            = "kubesphere.io/ks-version"
	InstallationModeAnnotation     = "kubesphere.io/installation-mode"
	ExternalDependenciesAnnotation = "kubesphere.io/external-dependencies"

	ExtensionReferenceLabel  = "kubesphere.io/extension-ref"
	RepositoryReferenceLabel = "kubesphere.io/repository-ref"
	CategoryLabel            = "kubesphere.io/category"

	ForceDeleteAnnotation = "kubesphere.io/force-delete"
)

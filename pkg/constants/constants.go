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
package constants

const (
	APIVersion = "v1alpha1"

	KubeSystemNamespace           = "kube-system"
	OpenPitrixNamespace           = "openpitrix-system"
	KubesphereDevOpsNamespace     = "kubesphere-devops-system"
	IstioNamespace                = "istio-system"
	KubeSphereMonitoringNamespace = "kubesphere-monitoring-system"
	KubeSphereLoggingNamespace    = "kubesphere-logging-system"
	KubeSphereNamespace           = "kubesphere-system"
	KubeSphereControlNamespace    = "kubesphere-controls-system"
	IngressControllerNamespace    = KubeSphereControlNamespace
	AdminUserName                 = "admin"
	DataHome                      = "/etc/kubesphere"
	IngressControllerFolder       = DataHome + "/ingress-controller"
	IngressControllerPrefix       = "kubesphere-router-"

	WorkspaceLabelKey              = "kubesphere.io/workspace"
	DisplayNameAnnotationKey       = "displayName"
	DescriptionAnnotationKey       = "desc"
	CreatorAnnotationKey           = "creator"
	System                         = "system"
	OpenPitrixRuntimeAnnotationKey = "openpitrix_runtime"
	WorkspaceAdmin                 = "workspace-admin"
	ClusterAdmin                   = "cluster-admin"
	WorkspaceRegular               = "workspace-regular"
	WorkspaceViewer                = "workspace-viewer"
	DevopsOwner                    = "owner"
	DevopsReporter                 = "reporter"

	UserNameHeader = "X-Token-Username"

	TenantResourcesTag         = "Tenant Resources"
	IdentityManagementTag      = "Identity Management"
	AccessManagementTag        = "Access Management"
	NamespaceResourcesTag      = "Namespace Resources"
	ClusterResourcesTag        = "Cluster Resources"
	ComponentStatusTag         = "Component Status"
	VerificationTag            = "Verification"
	RegistryTag                = "Docker Registry"
	UserResourcesTag           = "User Resources"
	DevOpsProjectTag           = "DevOps Project"
	DevOpsProjectCredentialTag = "DevOps Project Credential"
	DevOpsProjectMemberTag     = "DevOps Project Member"
	DevOpsPipelineTag          = "DevOps Pipeline"
	DevOpsWebhookTag           = "DevOps Webhook"
	DevOpsJenkinsfileTag       = "DevOps Jenkinsfile"
	DevOpsScmTag               = "DevOps Scm"
	ClusterMetricsTag          = "Cluster Metrics"
	NodeMetricsTag             = "Node Metrics"
	NamespaceMetricsTag        = "Namespace Metrics"
	PodMetricsTag              = "Pod Metrics"
	ContainerMetricsTag        = "Container Metrics"
	WorkloadMetricsTag         = "Workload Metrics"
	WorkspaceMetricsTag        = "Workspace Metrics"
	ComponentMetricsTag        = "Component Metrics"
	LogQueryTag                = "Log Query"
	FluentBitSetting           = "Fluent Bit Setting"
)

var (
	WorkSpaceRoles   = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
	SystemNamespaces = []string{KubeSphereNamespace, KubeSphereLoggingNamespace, KubeSphereMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, KubesphereDevOpsNamespace}
)

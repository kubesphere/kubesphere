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
	PorterNamespace               = "porter-system"
	IngressControllerNamespace    = KubeSphereControlNamespace
	AdminUserName                 = "admin"
	IngressControllerPrefix       = "kubesphere-router-"

	WorkspaceLabelKey        = "kubesphere.io/workspace"
	NamespaceLabelKey        = "kubesphere.io/namespace"
	DisplayNameAnnotationKey = "kubesphere.io/alias-name"
	DescriptionAnnotationKey = "kubesphere.io/description"
	CreatorAnnotationKey     = "kubesphere.io/creator"
	UsernameLabelKey         = "kubesphere.io/username"
	DevOpsProjectLabelKey    = "kubesphere.io/devopsproject"
	KubefedManagedLabel      = "kubefed.io/managed"

	UserNameHeader = "X-Token-Username"

	AuthenticationTag = "Authentication"
	UserTag           = "User"
	GroupTag          = "Group"

	WorkspaceMemberTag     = "Workspace Member"
	DevOpsProjectMemberTag = "DevOps Project Member"
	NamespaceMemberTag     = "Namespace Member"
	ClusterMemberTag       = "Cluster Member"

	GlobalRoleTag        = "Global Role"
	ClusterRoleTag       = "Cluster Role"
	WorkspaceRoleTag     = "Workspace Role"
	DevOpsProjectRoleTag = "DevOps Project Role"
	NamespaceRoleTag     = "Namespace Role"

	OpenpitrixAppInstanceTag = "App Instance"
	OpenpitrixAppTemplateTag = "App Template"
	OpenpitrixCategoryTag    = "Category"
	OpenpitrixAttachmentTag  = "Attachment"
	OpenpitrixRepositoryTag  = "Repository"
	OpenpitrixManagementTag  = "App Management"

	DevOpsCredentialTag  = "DevOps Credential"
	DevOpsPipelineTag    = "DevOps Pipeline"
	DevOpsWebhookTag     = "DevOps Webhook"
	DevOpsJenkinsfileTag = "DevOps Jenkinsfile"
	DevOpsScmTag         = "DevOps Scm"
	DevOpsJenkinsTag     = "Jenkins"

	ToolboxTag      = "Toolbox"
	RegistryTag     = "Docker Registry"
	GitTag          = "Git"
	TerminalTag     = "Terminal"
	MultiClusterTag = "Multi-cluster"

	WorkspaceTag     = "Workspace"
	NamespaceTag     = "Namespace"
	DevOpsProjectTag = "DevOps Project"
	UserResourceTag  = "User's Resources"

	NamespaceResourcesTag = "Namespace Resources"
	ClusterResourcesTag   = "Cluster Resources"
	ComponentStatusTag    = "Component Status"

	NetworkTopologyTag = "Network Topology"

	KubeSphereMetricsTag = "KubeSphere Metrics"
	ClusterMetricsTag    = "Cluster Metrics"
	NodeMetricsTag       = "Node Metrics"
	NamespaceMetricsTag  = "Namespace Metrics"
	PodMetricsTag        = "Pod Metrics"
	PVCMetricsTag        = "PVC Metrics"
	ContainerMetricsTag  = "Container Metrics"
	WorkloadMetricsTag   = "Workload Metrics"
	WorkspaceMetricsTag  = "Workspace Metrics"
	ComponentMetricsTag  = "Component Metrics"
	CustomMetricsTag     = "Custom Metrics"

	LogQueryTag      = "Log Query"
	EventsQueryTag   = "Events Query"
	AuditingQueryTag = "Auditing Query"
)

var (
	SystemNamespaces = []string{KubeSphereNamespace, KubeSphereLoggingNamespace, KubeSphereMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, KubesphereDevOpsNamespace, PorterNamespace}
)

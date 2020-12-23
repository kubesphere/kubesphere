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

	ClusterNameLabelKey               = "kubesphere.io/cluster"
	NameLabelKey                      = "kubesphere.io/name"
	WorkspaceLabelKey                 = "kubesphere.io/workspace"
	NamespaceLabelKey                 = "kubesphere.io/namespace"
	DisplayNameAnnotationKey          = "kubesphere.io/alias-name"
	ChartRepoIdLabelKey               = "kubesphere.io/repo-id"
	ChartApplicationIdLabelKey        = "kubesphere.io/app-id"
	ChartApplicationVersionIdLabelKey = "kubesphere.io/ver-id"
	CategoryIdLabelKey                = "kubesphere.io/ctg-id"
	CreatorAnnotationKey              = "kubesphere.io/creator"
	UsernameLabelKey                  = "kubesphere.io/username"
	DevOpsProjectLabelKey             = "kubesphere.io/devopsproject"
	KubefedManagedLabel               = "kubefed.io/managed"

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

	OpenpitrixTag            = "Openpitrix Resources"
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

	MsgLen               = 512
	HelmRepoSyncStateLen = 10
	HelmAuditLen         = 10

	//app version state
	StateDraft     = "draft"
	StateSubmitted = "submitted"
	StatePassed    = "passed"
	StateRejected  = "rejected"
	StateSuspended = "suspended"
	StateActive    = "active"

	HelmStatusActive      = "active"
	HelmStatusUsed        = "used"
	HelmStatusEnabled     = "enabled"
	HelmStatusDisabled    = "disabled"
	HelmStatusCreating    = "creating"
	HelmStatusDeleted     = "deleted"
	HelmStatusDeleting    = "deleting"
	HelmStatusUpgrading   = "upgrading"
	HelmStatusUpdating    = "updating"
	HelmStatusRollbacking = "rollbacking"
	HelmStatusStopped     = "stopped"
	HelmStatusStopping    = "stopping"
	HelmStatusStarting    = "starting"
	HelmStatusRecovering  = "recovering"
	HelmStatusCeased      = "ceased"
	HelmStatusCeasing     = "ceasing"
	HelmStatusResizing    = "resizing"
	HelmStatusScaling     = "scaling"
	HelmStatusWorking     = "working"
	HelmStatusPending     = "pending"
	HelmStatusSuccessful  = "successful"
	HelmStatusFailed      = "failed"

	AttachmentTypeScreenshot = "screenshot"
	AttachmentTypeIcon       = "icon"

	StateInReview = "in-review"
	StateNew      = "new"

	HelmApplicationAppStoreSuffix  = "-store"
	HelmApplicationIdPrefix        = "app-"
	HelmRepoIdPrefix               = "repo-"
	HelmApplicationVersionIdPrefix = "appv-"
	HelmCategoryIdPrefix           = "ctg-"
	HelmAttachmentPrefix           = "att-"
	HelmReleasePrefix              = "rls-"
	UncategorizedName              = "uncategorized"
	UncategorizedId                = "ctg-uncategorized"
	AppStoreRepoId                 = "repo-helm"
)

var (
	SystemNamespaces = []string{KubeSphereNamespace, KubeSphereLoggingNamespace, KubeSphereMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, KubesphereDevOpsNamespace, PorterNamespace}
)

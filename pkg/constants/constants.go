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

	WorkspaceLabelKey              = "kubesphere.io/workspace"
	NamespaceLabelKey              = "kubesphere.io/namespace"
	RuntimeLabelKey                = "openpitrix.io/namespace"
	DisplayNameAnnotationKey       = "kubesphere.io/alias-name"
	DescriptionAnnotationKey       = "kubesphere.io/description"
	CreatorAnnotationKey           = "kubesphere.io/creator"
	UsernameLabelKey               = "kubesphere.io/username"
	System                         = "system"
	OpenPitrixRuntimeAnnotationKey = "openpitrix_runtime"
	WorkspaceAdmin                 = "workspace-admin"
	ClusterAdmin                   = "cluster-admin"
	WorkspaceRegular               = "workspace-regular"
	WorkspaceViewer                = "workspace-viewer"
	WorkspacesManager              = "workspaces-manager"
	DevopsOwner                    = "owner"
	DevopsReporter                 = "reporter"
	DevOpsProjectLabelKey          = "kubesphere.io/devopsproject"
	KubefedManagedLabel            = "kubefed.io/managed"

	UserNameHeader = "X-Token-Username"

	TenantResourcesTag         = "Tenant Resources"
	IdentityManagementTag      = "Identity Management"
	AccessManagementTag        = "Access Management"
	NamespaceResourcesTag      = "Namespace Resources"
	ClusterResourcesTag        = "Cluster Resources"
	ComponentStatusTag         = "Component Status"
	OpenpitrixTag              = "Openpitrix Resources"
	VerificationTag            = "Verification"
	RegistryTag                = "Docker Registry"
	NetworkTopologyTag         = "Network Topology"
	UserResourcesTag           = "User Resources"
	DevOpsProjectTag           = "DevOps Project"
	DevOpsProjectCredentialTag = "DevOps Project Credential"
	DevOpsProjectMemberTag     = "DevOps Project Member"
	DevOpsPipelineTag          = "DevOps Pipeline"
	DevOpsWebhookTag           = "DevOps Webhook"
	DevOpsJenkinsfileTag       = "DevOps Jenkinsfile"
	DevOpsScmTag               = "DevOps Scm"
	KubeSphereMetricsTag       = "KubeSphere Metrics"
	ClusterMetricsTag          = "Cluster Metrics"
	NodeMetricsTag             = "Node Metrics"
	NamespaceMetricsTag        = "Namespace Metrics"
	PodMetricsTag              = "Pod Metrics"
	PVCMetricsTag              = "PVC Metrics"
	ContainerMetricsTag        = "Container Metrics"
	WorkloadMetricsTag         = "Workload Metrics"
	WorkspaceMetricsTag        = "Workspace Metrics"
	ComponentMetricsTag        = "Component Metrics"
	CustomMetricsTag           = "Custom Metrics"
	LogQueryTag                = "Log Query"
	TerminalTag                = "Terminal"
	EventsQueryTag             = "Events Query"
	AuditingQueryTag           = "Auditing Query"
)

var (
	WorkSpaceRoles   = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
	SystemNamespaces = []string{KubeSphereNamespace, KubeSphereLoggingNamespace, KubeSphereMonitoringNamespace, OpenPitrixNamespace, KubeSystemNamespace, IstioNamespace, KubesphereDevOpsNamespace, PorterNamespace}
)

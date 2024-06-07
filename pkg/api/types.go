/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ListResult struct {
	Items      []runtime.Object `json:"items"`
	TotalItems int              `json:"totalItems"`
}

type ResourceQuota struct {
	Namespace string                     `json:"namespace" description:"namespace"`
	Data      corev1.ResourceQuotaStatus `json:"data" description:"resource quota status"`
}

type NamespacedResourceQuota struct {
	Namespace string `json:"namespace,omitempty"`

	Data struct {
		corev1.ResourceQuotaStatus

		// quota left status, do the math on the side, cause it's
		// a lot easier with go-client library
		Left corev1.ResourceList `json:"left,omitempty"`
	} `json:"data,omitempty"`
}

type Router struct {
	RouterType  string            `json:"type"`
	Annotations map[string]string `json:"annotations"`
}

type GitCredential struct {
	RemoteUrl string                  `json:"remoteUrl" description:"git server url"`
	SecretRef *corev1.SecretReference `json:"secretRef,omitempty" description:"auth secret reference"`
}

type RegistryCredential struct {
	Username   string `json:"username" description:"username"`
	Password   string `json:"password" description:"password"`
	ServerHost string `json:"serverhost" description:"registry server host"`
}

type Workloads struct {
	Namespace string                 `json:"namespace" description:"the name of the namespace"`
	Count     map[string]int         `json:"data" description:"the number of unhealthy workloads"`
	Items     map[string]interface{} `json:"items,omitempty" description:"unhealthy workloads"`
}

const (
	ResourceKindDaemonSet             = "daemonsets"
	ResourceKindDeployment            = "deployments"
	ResourceKindJob                   = "jobs"
	ResourceKindPersistentVolumeClaim = "persistentvolumeclaims"
	ResourceKindStatefulSet           = "statefulsets"
	StatusOK                          = "ok"
	WorkspaceNone                     = ""
	ClusterNone                       = ""
	TagNonResourceAPI                 = "NonResource APIs"
	TagAuthentication                 = "Authentication"
	TagMultiCluster                   = "Multi-cluster"
	TagIdentityManagement             = "Identity Management"
	TagAccessManagement               = "Access Management"
	TagAdvancedOperations             = "Advanced Operations"
	TagTerminal                       = "Web Terminal"
	TagNamespacedResources            = "Namespaced Resources"
	TagClusterResources               = "Cluster Resources"
	TagComponentStatus                = "Component Status"
	TagUserRelatedResources           = "User Related Resources"
	TagPlatformConfigurations         = "Platform Configurations"
)

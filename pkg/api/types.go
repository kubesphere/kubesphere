/*
Copyright 2020 The KubeSphere Authors.

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

package api

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ListResult struct {
	Items      []interface{} `json:"items"`
	TotalItems int           `json:"totalItems"`
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

type ClientType string

const (
	ClientKubernetes  ClientType = "Kubernetes"
	ClientKubeSphere  ClientType = "Kubesphere"
	ClientIstio       ClientType = "Istio"
	ClientS2i         ClientType = "S2i"
	ClientApplication ClientType = "Application"

	StatusOK = "ok"
)

var SupportedGroupVersionResources = map[ClientType][]schema.GroupVersionResource{
	// all supported kubernetes api objects
	ClientKubernetes: {
		{Group: "", Version: "v1", Resource: "namespaces"},
		{Group: "", Version: "v1", Resource: "nodes"},
		{Group: "", Version: "v1", Resource: "resourcequotas"},
		{Group: "", Version: "v1", Resource: "pods"},
		{Group: "", Version: "v1", Resource: "services"},
		{Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
		{Group: "", Version: "v1", Resource: "secrets"},
		{Group: "", Version: "v1", Resource: "configmaps"},
		{Group: "", Version: "v1", Resource: "serviceaccounts"},

		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
		{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},

		{Group: "apps", Version: "v1", Resource: "deployments"},
		{Group: "apps", Version: "v1", Resource: "daemonsets"},
		{Group: "apps", Version: "v1", Resource: "replicasets"},
		{Group: "apps", Version: "v1", Resource: "statefulsets"},
		{Group: "apps", Version: "v1", Resource: "controllerrevisions"},

		{Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},

		{Group: "batch", Version: "v1", Resource: "jobs"},
		{Group: "batch", Version: "v1beta1", Resource: "cronjobs"},

		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},

		{Group: "autoscaling", Version: "v2beta2", Resource: "horizontalpodautoscalers"},
	},

	// all supported kubesphere api objects
	ClientKubeSphere: {
		{Group: "tenant.kubesphere.io", Version: "v1alpha1", Resource: "workspaces"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibinaries"},

		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "strategies"},
		{Group: "servicemesh.kubesphere.io", Version: "v1alpha2", Resource: "servicepolicies"},
	},

	// all supported istio api objects
	ClientIstio: {},

	// all supported s2i api objects
	// TODO: move s2i clientset into kubesphere
	ClientS2i: {
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuildertemplates"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2iruns"},
		{Group: "devops.kubesphere.io", Version: "v1alpha1", Resource: "s2ibuilders"},
	},

	// kubernetes-sigs application api objects
	ClientApplication: {
		{Group: "app.k8s.io", Version: "v1beta1", Resource: "applications"},
	},
}

// List of all resource kinds supported by the UI.
const (
	ResourceKindConfigMap                = "configmaps"
	ResourceKindDaemonSet                = "daemonsets"
	ResourceKindDeployment               = "deployments"
	ResourceKindEvent                    = "events"
	ResourceKindHorizontalPodAutoscaler  = "horizontalpodautoscalers"
	ResourceKindIngress                  = "ingresses"
	ResourceKindJob                      = "jobs"
	ResourceKindCronJob                  = "cronjobs"
	ResourceKindLimitRange               = "limitranges"
	ResourceKindNamespace                = "namespaces"
	ResourceKindNode                     = "nodes"
	ResourceKindPersistentVolumeClaim    = "persistentvolumeclaims"
	ResourceKindPersistentVolume         = "persistentvolumes"
	ResourceKindCustomResourceDefinition = "customresourcedefinitions"
	ResourceKindPod                      = "pods"
	ResourceKindReplicaSet               = "replicasets"
	ResourceKindResourceQuota            = "resourcequota"
	ResourceKindSecret                   = "secrets"
	ResourceKindService                  = "services"
	ResourceKindStatefulSet              = "statefulsets"
	ResourceKindStorageClass             = "storageclasses"
	ResourceKindClusterRole              = "clusterroles"
	ResourceKindClusterRoleBinding       = "clusterrolebindings"
	ResourceKindRole                     = "roles"
	ResourceKindRoleBinding              = "rolebindings"
	ResourceKindWorkspace                = "workspaces"
	ResourceKindS2iBinary                = "s2ibinaries"
	ResourceKindStrategy                 = "strategy"
	ResourceKindServicePolicy            = "servicepolicies"
	ResourceKindS2iBuilderTemplate       = "s2ibuildertemplates"
	ResourceKindeS2iRun                  = "s2iruns"
	ResourceKindS2iBuilder               = "s2ibuilders"
	ResourceKindApplication              = "applications"

	WorkspaceNone = ""
	ClusterNone   = ""
)

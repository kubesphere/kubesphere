/*
Copyright 2018 The KubeSphere Authors.

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

import "os"

type MessageResponse struct {
	Message string `json:"message"`
}

type PageableResponse struct {
	Items      []interface{} `json:"items"`
	TotalCount int           `json:"total_count"`
}

const (
	APIVersion = "v1alpha1"

	KubeSystemNamespace        = "kube-system"
	OpenPitrixNamespace        = "openpitrix-system"
	IstioNamespace             = "istio-system"
	KubeSphereNamespace        = "kubesphere-system"
	KubeSphereControlNamespace = "kubesphere-controls-system"
	IngressControllerNamespace = KubeSphereControlNamespace
	AdminUserName              = "admin"
	DataHome                   = "/etc/kubesphere"
	IngressControllerFolder    = DataHome + "/ingress-controller"
	IngressControllerPrefix    = "kubesphere-router-"
	DevopsAPIServerEnv         = "DEVOPS_API_SERVER"
	AccountAPIServerEnv        = "ACCOUNT_API_SERVER"
	DevopsProxyTokenEnv        = "DEVOPS_PROXY_TOKEN"
	OpenPitrixProxyTokenEnv    = "OPENPITRIX_PROXY_TOKEN"
	WorkspaceLabelKey          = "kubesphere.io/workspace"
	WorkspaceAdmin             = "workspace-admin"
	ClusterAdmin               = "cluster-admin"
	WorkspaceRegular           = "workspace-regular"
	WorkspaceViewer            = "workspace-viewer"
	DevopsOwner                = "owner"
	DevopsReporter             = "reporter"
)

var (
	DevopsAPIServer      = "ks-devops-apiserver.kubesphere-system.svc"
	AccountAPIServer     = "ks-account.kubesphere-system.svc"
	DevopsProxyToken     = ""
	OpenPitrixProxyToken = ""
	WorkSpaceRoles       = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
)

type LogQueryLevel int

const (
	QueryLevelCluster LogQueryLevel = iota
	QueryLevelWorkspace
	QueryLevelNamespace
	QueryLevelWorkload
	QueryLevelPod
	QueryLevelContainer
)

const (
	LogQueryLevelCluster   = "Cluster"
	LogQueryLevelWorkspace = "Workspace"
	LogQueryLevelNamespace = "Namespace"
	LogQueryLevelWorkload  = "Workload"
	LogQueryLevelPod       = "Pod"
	LogQueryLevelContainer = "Container"
)

type FormatedLevelLog struct {
	LogLevel string `json:"log_query_level"`
}

const (
	LogQueryOperationQuery      = "Query"
	LogQueryOperationStatistics = "Statistics"
)

func init() {
	if env := os.Getenv(DevopsAPIServerEnv); env != "" {
		DevopsAPIServer = env
	}
	if env := os.Getenv(AccountAPIServerEnv); env != "" {
		AccountAPIServer = env
	}
	if env := os.Getenv(DevopsProxyTokenEnv); env != "" {
		DevopsProxyToken = env
	}
	if env := os.Getenv(OpenPitrixProxyTokenEnv); env != "" {
		OpenPitrixProxyToken = env
	}
}

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

	WorkspaceLabelKey = "kubesphere.io/workspace"
	WorkspaceAdmin    = "workspace-admin"
	ClusterAdmin      = "cluster-admin"
	WorkspaceRegular  = "workspace-regular"
	WorkspaceViewer   = "workspace-viewer"
	DevopsOwner       = "owner"
	DevopsReporter    = "reporter"

	envDevopsAPIServer      = "DEVOPS_API_SERVER"
	envAccountAPIServer     = "ACCOUNT_API_SERVER"
	envDevopsProxyToken     = "DEVOPS_PROXY_TOKEN"
	envOpenPitrixProxyToken = "OPENPITRIX_PROXY_TOKEN"

	UserNameHeader = "X-Token-Username"
)

var (
	WorkSpaceRoles   = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
	SystemWorkspace  = "system-workspace"
	DevopsAPIServer  = "ks-devops-apiserver.kubesphere-system.svc"
	AccountAPIServer = "ks-account.kubesphere-system.svc"
	SystemNamespaces = []string{KubeSphereNamespace, OpenPitrixNamespace, KubeSystemNamespace}
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

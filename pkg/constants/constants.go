package constants

import "os"

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

	DevopsAPIServerEnv      = "DEVOPS_API_SERVER"
	AccountAPIServerEnv     = "ACCOUNT_API_SERVER"
	DevopsProxyTokenEnv     = "DEVOPS_PROXY_TOKEN"
	OpenPitrixProxyTokenEnv = "OPENPITRIX_PROXY_TOKEN"
)

var (
	WorkSpaceRoles       = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
	DevopsAPIServer      = "ks-devops-apiserver.kubesphere-system.svc"
	AccountAPIServer     = "ks-account.kubesphere-system.svc"
	DevopsProxyToken     = ""
	OpenPitrixProxyToken = ""
	SystemNamespaces     = []string{KubeSystemNamespace, OpenPitrixNamespace, KubeSystemNamespace}
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

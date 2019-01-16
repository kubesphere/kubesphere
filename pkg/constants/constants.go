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

	envDevopsAPIServer      = "DEVOPS_API_SERVER"
	envAccountAPIServer     = "ACCOUNT_API_SERVER"
	envDevopsProxyToken     = "DEVOPS_PROXY_TOKEN"
	envOpenPitrixProxyToken = "OPENPITRIX_PROXY_TOKEN"
	envRootDN               = "ROOT_DN"
	envRootPWD              = "ROOT_PWD"
	envUserSearchBase       = "USER_SEARCH_BASE"
	envGroupSearchBase      = "GROUP_SEARCH_BASE"
	envLDAPServerHost       = "LDAP_SERVER_HOST"
	envAdminEmail           = "ADMIN_EMAIL"
	envAdminPWD             = "ADMIN_PWD"

	UserNameHeader = "X-Token-Username"
)

var (
	AdminEmail           = "admin@kubesphere.io"
	AdminPWD             = "passw0rd"
	RootDN               = "cn=admin,dc=example,dc=org"
	RootPWD              = "admin"
	UserSearchBase       = "ou=Users,dc=example,dc=org"
	GroupSearchBase      = "ou=Groups,dc=example,dc=org"
	LdapServerHost       = "localhost:389"
	WorkSpaceRoles       = []string{WorkspaceAdmin, WorkspaceRegular, WorkspaceViewer}
	SystemWorkspace      = "system-workspace"
	DevopsAPIServer      = "ks-devops-apiserver.kubesphere-system.svc"
	AccountAPIServer     = "ks-account.kubesphere-system.svc"
	DevopsProxyToken     = ""
	OpenPitrixProxyToken = ""
	SystemNamespaces     = []string{KubeSystemNamespace, OpenPitrixNamespace, KubeSystemNamespace}
)

func init() {
	if env := os.Getenv(envDevopsAPIServer); env != "" {
		DevopsAPIServer = env
	}
	if env := os.Getenv(envAccountAPIServer); env != "" {
		AccountAPIServer = env
	}
	if env := os.Getenv(envDevopsProxyToken); env != "" {
		DevopsProxyToken = env
	}
	if env := os.Getenv(envOpenPitrixProxyToken); env != "" {
		OpenPitrixProxyToken = env
	}
	if env := os.Getenv(envRootDN); env != "" {
		RootDN = env
	}
	if env := os.Getenv(envRootPWD); env != "" {
		RootPWD = env
	}
	if env := os.Getenv(envUserSearchBase); env != "" {
		UserSearchBase = env
	}
	if env := os.Getenv(envGroupSearchBase); env != "" {
		GroupSearchBase = env
	}
	if env := os.Getenv(envLDAPServerHost); env != "" {
		LdapServerHost = env
	}
	if env := os.Getenv(envAdminEmail); env != "" {
		AdminEmail = env
	}
	if env := os.Getenv(envAdminPWD); env != "" {
		AdminPWD = env
	}
}

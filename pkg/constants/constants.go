/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package constants

const (
	KubeSystemNamespace        = "kube-system"
	KubeSphereNamespace        = "kubesphere-system"
	KubeSphereAPIServerName    = "ks-apiserver"
	KubeSphereConfigName       = "kubesphere-config"
	KubeSphereConfigMapDataKey = "kubesphere.yaml"
	KubectlPodNamePrefix       = "ks-managed-kubectl"

	WorkspaceLabelKey        = "kubesphere.io/workspace"
	DisplayNameAnnotationKey = "kubesphere.io/alias-name"
	DescriptionAnnotationKey = "kubesphere.io/description"
	CreatorAnnotationKey     = "kubesphere.io/creator"
	UsernameLabelKey         = "kubesphere.io/username"
	GenericConfigTypeLabel   = "config.kubesphere.io/type"
	KubectlPodLabel          = "kubesphere.io/kubectl-pod"
	ConfigHashAnnotation     = "kubesphere.io/config-hash"
	KubeSphereManagedLabel   = "kubesphere.io/managed"
)

var (
	SystemNamespaces = []string{KubeSphereNamespace, KubeSystemNamespace}
)

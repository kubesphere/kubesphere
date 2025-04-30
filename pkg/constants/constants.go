/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package constants

import corev1 "k8s.io/api/core/v1"

const (
	SystemWorkspace            = "system-workspace"
	KubeSystemNamespace        = "kube-system"
	KubeSphereNamespace        = "kubesphere-system"
	KubeSphereAPIServerName    = "ks-apiserver"
	KubeSphereConfigName       = "kubesphere-config"
	KubeSphereConfigMapDataKey = "kubesphere.yaml"
	KubectlPodNamePrefix       = "ks-managed-kubectl"

	ProtectedResourceLabel        = "kubesphere.io/protected-resource"
	WorkspaceLabelKey             = "kubesphere.io/workspace"
	DisplayNameAnnotationKey      = "kubesphere.io/alias-name"
	DescriptionAnnotationKey      = "kubesphere.io/description"
	CreatorAnnotationKey          = "kubesphere.io/creator"
	UsernameLabelKey              = "kubesphere.io/username"
	GenericConfigTypeLabel        = "config.kubesphere.io/type"
	KubectlPodLabel               = "kubesphere.io/kubectl-pod"
	ConfigHashAnnotation          = "kubesphere.io/config-hash"
	KubeSphereManagedLabel        = "kubesphere.io/managed"
	DeletionPropagationAnnotation = "kubesphere.io/deletion-propagation"
	CascadingDeletionFinalizer    = "kubesphere.io/cascading-deletion"

	KubeSphereConfigGroup                             = "config.kubesphere.io"
	SecretTypeGenericPlatformConfig corev1.SecretType = KubeSphereConfigGroup + "/generic-platform-config"
	GenericPlatformConfigNameFmt                      = "io.kubesphere.config.platformconfig.%s"
	GenericPlatformConfigFileName                     = "configuration.yaml"

	SecretTypeClusterConnectionConfig corev1.SecretType = KubeSphereConfigGroup + "/cluster-connection-config"
	ClusterConnectionConfigFileName                     = "configuration.yaml"
)

var (
	SystemNamespaces = []string{KubeSphereNamespace, KubeSystemNamespace}
)

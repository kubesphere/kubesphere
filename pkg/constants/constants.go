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

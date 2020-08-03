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

// NOTE: Boilerplate only. Ignore this file.

// +k8s:openapi-gen=true
// +k8s:deepcopy-gen=package,register
// +k8s:conversion-gen=kubesphere.io/kubesphere/pkg/apis/types
// +k8s:defaulter-gen=TypeMeta
// +groupName=types.kubefed.io
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "types.kubefed.io", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme is required by pkg/client/...
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func init() {
	SchemeBuilder.Register(
		&FederatedApplication{},
		&FederatedApplicationList{},
		&FederatedClusterRole{},
		&FederatedClusterRoleList{},
		&FederatedClusterRoleBinding{},
		&FederatedClusterRoleBindingList{},
		&FederatedConfigMap{},
		&FederatedConfigMapList{},
		&FederatedDeployment{},
		&FederatedDeploymentList{},
		&FederatedIngress{},
		&FederatedIngressList{},
		&FederatedLimitRange{},
		&FederatedLimitRangeList{},
		&FederatedNamespace{},
		&FederatedNamespaceList{},
		&FederatedPersistentVolumeClaim{},
		&FederatedPersistentVolumeClaimList{},
		&FederatedResourceQuota{},
		&FederatedResourceQuotaList{},
		&FederatedSecret{},
		&FederatedSecretList{},
		&FederatedService{},
		&FederatedServiceList{},
		&FederatedStatefulSet{},
		&FederatedStatefulSetList{},
		&FederatedUser{},
		&FederatedUserList{},
		&FederatedWorkspace{},
		&FederatedWorkspaceList{})
}

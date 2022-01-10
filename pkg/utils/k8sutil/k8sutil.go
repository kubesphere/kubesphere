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

package k8sutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	tenantv1alpha2 "kubesphere.io/api/tenant/v1alpha2"
)

// IsControlledBy returns whether the ownerReferences contains the specified resource kind
func IsControlledBy(ownerReferences []metav1.OwnerReference, kind string, name string) bool {
	for _, owner := range ownerReferences {
		if owner.Kind == kind && (name == "" || owner.Name == name) {
			return true
		}
	}
	return false
}

// RemoveWorkspaceOwnerReference remove workspace kind owner reference
func RemoveWorkspaceOwnerReference(ownerReferences []metav1.OwnerReference) []metav1.OwnerReference {
	tmp := make([]metav1.OwnerReference, 0)
	for _, owner := range ownerReferences {
		if owner.Kind != tenantv1alpha1.ResourceKindWorkspace &&
			owner.Kind != tenantv1alpha2.ResourceKindWorkspaceTemplate {
			tmp = append(tmp, owner)
		}
	}
	return tmp
}

// GetWorkspaceOwnerName return workspace kind owner name
func GetWorkspaceOwnerName(ownerReferences []metav1.OwnerReference) string {
	for _, owner := range ownerReferences {
		if owner.Kind == tenantv1alpha1.ResourceKindWorkspace ||
			owner.Kind == tenantv1alpha2.ResourceKindWorkspaceTemplate {
			return owner.Name
		}
	}
	return ""
}

// LoadKubeConfigFromBytes parses the kubeconfig yaml data to the rest.Config struct.
func LoadKubeConfigFromBytes(kubeconfig []byte) (*rest.Config, error) {
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	return config, nil
}

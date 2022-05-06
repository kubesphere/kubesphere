/*
Copyright 2020 KubeSphere Authors

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

package cluster

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// HostClientset provides a kubernetes API compliant clientset to
// communicate with the host cluster's kubernetes API server.
func HostClientset(config *rest.Config) (*kubeclient.Clientset, error) {
	return kubeclient.NewForConfig(config)
}

// ClusterClientset provides a kubernetes API compliant clientset to
// communicate with the joining cluster's kubernetes API server.
func ClusterClientset(config *rest.Config) (*kubeclient.Clientset, error) {
	return kubeclient.NewForConfig(config)
}

// ClusterServiceAccountName returns the name of a service account whose
// credentials are used by the host cluster to access the client cluster.
func ClusterServiceAccountName(joiningClusterName, hostClusterName string) string {
	return fmt.Sprintf("%s-%s", joiningClusterName, hostClusterName)
}

// IsPrimaryCluster checks if the caller is working with objects for the
// primary cluster by checking if the UIDs match for both ObjectMetas passed
// in.
// TODO (font): Need to revisit this when cluster ID is available.
func IsPrimaryCluster(obj, clusterObj pkgruntime.Object) bool {
	meta := MetaAccessor(obj)
	clusterMeta := MetaAccessor(clusterObj)
	return meta.GetUID() == clusterMeta.GetUID()
}

func MetaAccessor(obj pkgruntime.Object) metav1.Object {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		// This should always succeed if obj is not nil.  Also,
		// adapters are slated for replacement by unstructured.
		return nil
	}
	return accessor
}

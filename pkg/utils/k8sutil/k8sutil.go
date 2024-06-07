/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package k8sutil

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
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
		if owner.Kind != tenantv1beta1.ResourceKindWorkspace &&
			owner.Kind != tenantv1beta1.ResourceKindWorkspaceTemplate {
			tmp = append(tmp, owner)
		}
	}
	return tmp
}

// GetWorkspaceOwnerName return workspace kind owner name
func GetWorkspaceOwnerName(ownerReferences []metav1.OwnerReference) string {
	for _, owner := range ownerReferences {
		if owner.Kind == tenantv1beta1.ResourceKindWorkspace ||
			owner.Kind == tenantv1beta1.ResourceKindWorkspaceTemplate {
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

func GetObjectMeta(obj metav1.Object) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:                       obj.GetName(),
		GenerateName:               obj.GetGenerateName(),
		Namespace:                  obj.GetNamespace(),
		UID:                        obj.GetUID(),
		ResourceVersion:            obj.GetResourceVersion(),
		Generation:                 obj.GetGeneration(),
		CreationTimestamp:          obj.GetCreationTimestamp(),
		DeletionTimestamp:          obj.GetDeletionTimestamp(),
		DeletionGracePeriodSeconds: obj.GetDeletionGracePeriodSeconds(),
		Labels:                     obj.GetLabels(),
		Annotations:                obj.GetAnnotations(),
		OwnerReferences:            obj.GetOwnerReferences(),
		Finalizers:                 obj.GetFinalizers(),
		ManagedFields:              obj.GetManagedFields(),
	}
}

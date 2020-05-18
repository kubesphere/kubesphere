/*
Copyright 2019 The Kubernetes Authors.

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

package util

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

const (
	// If this annotation is present on a federated resource, resources in the
	// member clusters managed by the federated resource should be orphaned.
	// If the annotation is not present (the default), resources in member
	// clusters will be deleted before the federated resource is deleted.
	OrphanManagedResourcesAnnotation = "kubefed.io/orphan"
	OrphanedManagedResourcesValue    = "true"
)

// IsOrphaningEnabled checks status of "orphaning enable" (OrphanManagedResources: OrphanedManagedResourceslValue')
// annotation on a resource.
func IsOrphaningEnabled(obj *unstructured.Unstructured) bool {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return false
	}
	return annotations[OrphanManagedResourcesAnnotation] == OrphanedManagedResourcesValue
}

// Enables the orphaning mode
func EnableOrphaning(obj *unstructured.Unstructured) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[OrphanManagedResourcesAnnotation] = OrphanedManagedResourcesValue
	obj.SetAnnotations(annotations)
}

// Disables the orphaning mode
func DisableOrphaning(obj *unstructured.Unstructured) {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		return
	}
	delete(annotations, OrphanManagedResourcesAnnotation)
	obj.SetAnnotations(annotations)
}

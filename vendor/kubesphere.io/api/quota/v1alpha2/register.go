// NOTE: Boilerplate only. Ignore this file.

// Package v1alpha2 contains API Schema definitions for the quotas v1alpha2 API group
// +k8s:openapi-gen=true
// +kubebuilder:object:generate=true
// +k8s:conversion-gen=kubesphere.io/api/quota
// +k8s:defaulter-gen=TypeMeta
// +groupName=quota.kubesphere.io
package v1alpha2

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "quota.kubesphere.io", Version: "v1alpha2"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}

	// AddToScheme is required by pkg/client/...
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource is required by pkg/client/listers/...
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

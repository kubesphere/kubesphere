// NOTE: Boilerplate only. Ignore this file.

// +kubebuilder:object:generate=true
// +groupName=extensions.kubesphere.io

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: "extensions.kubesphere.io", Version: "v1alpha1"}

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
		&APIService{},
		&APIServiceList{},
		&JSBundle{},
		&JSBundleList{},
		&ReverseProxy{},
		&ReverseProxyList{},
		&ExtensionEntry{},
		&ExtensionEntryList{},
	)
}

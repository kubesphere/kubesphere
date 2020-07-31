package apis

import (
	typesv1beta1 "kubesphere.io/kubesphere/pkg/apis/types/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, typesv1beta1.AddToScheme)
}

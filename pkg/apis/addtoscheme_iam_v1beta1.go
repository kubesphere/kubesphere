package apis

import (
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, iamv1beta1.SchemeBuilder.AddToScheme)
}

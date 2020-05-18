package apis

import (
	"kubesphere.io/kubesphere/pkg/apis/cluster/v1alpha1"
)

func init() {
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
}

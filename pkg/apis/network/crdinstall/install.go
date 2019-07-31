package crdinstall

import (
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	networkv1alpha1 "kubesphere.io/kubesphere/pkg/apis/network/v1alpha1"
)

func Install(scheme *k8sruntime.Scheme) {
	urlruntime.Must(networkv1alpha1.AddToScheme(scheme))
	urlruntime.Must(scheme.SetVersionPriority(networkv1alpha1.SchemeGroupVersion))
}

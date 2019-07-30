package crdinstall

import (
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	urlruntime "k8s.io/apimachinery/pkg/util/runtime"
	servicemeshv1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
)

func Install(scheme *k8sruntime.Scheme) {
	urlruntime.Must(servicemeshv1alpha2.AddToScheme(scheme))
	urlruntime.Must(scheme.SetVersionPriority(servicemeshv1alpha2.SchemeGroupVersion))
}

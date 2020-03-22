package informers

import (
	appinformers "github.com/kubernetes-sigs/application/pkg/client/informers/externalversions"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	"k8s.io/client-go/informers"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

type nullInformerFactory struct {
}

func NewNullInformerFactory() InformerFactory {
	return &nullInformerFactory{}
}

func (n nullInformerFactory) KubernetesSharedInformerFactory() informers.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) IstioSharedInformerFactory() istioinformers.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) ApplicationSharedInformerFactory() appinformers.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) Start(stopCh <-chan struct{}) {
}

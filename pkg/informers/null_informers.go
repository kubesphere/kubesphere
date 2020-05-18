package informers

import (
	snapshotinformer "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	appinformers "github.com/kubernetes-sigs/application/pkg/client/informers/externalversions"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
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

func (n nullInformerFactory) SnapshotSharedInformerFactory() snapshotinformer.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) ApiExtensionSharedInformerFactory() apiextensionsinformers.SharedInformerFactory {
	return nil
}

func (n nullInformerFactory) Start(stopCh <-chan struct{}) {
}

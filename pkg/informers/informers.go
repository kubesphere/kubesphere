/*
Copyright 2019 The KubeSphere Authors.

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

package informers

import (
	snapshotclient "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/clientset/versioned"
	snapshotinformer "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	applicationclient "sigs.k8s.io/application/pkg/client/clientset/versioned"
	applicationinformers "sigs.k8s.io/application/pkg/client/informers/externalversions"
	"time"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which kubesphere needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory
	IstioSharedInformerFactory() istioinformers.SharedInformerFactory
	ApplicationSharedInformerFactory() applicationinformers.SharedInformerFactory
	SnapshotSharedInformerFactory() snapshotinformer.SharedInformerFactory
	ApiExtensionSharedInformerFactory() apiextensionsinformers.SharedInformerFactory

	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory              k8sinformers.SharedInformerFactory
	ksInformerFactory            ksinformers.SharedInformerFactory
	istioInformerFactory         istioinformers.SharedInformerFactory
	appInformerFactory           applicationinformers.SharedInformerFactory
	snapshotInformerFactory      snapshotinformer.SharedInformerFactory
	apiextensionsInformerFactory apiextensionsinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, ksClient versioned.Interface, istioClient istioclient.Interface,
	appClient applicationclient.Interface, snapshotClient snapshotclient.Interface, apiextensionsClient apiextensionsclient.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if ksClient != nil {
		factory.ksInformerFactory = ksinformers.NewSharedInformerFactory(ksClient, defaultResync)
	}

	if appClient != nil {
		factory.appInformerFactory = applicationinformers.NewSharedInformerFactory(appClient, defaultResync)
	}

	if istioClient != nil {
		factory.istioInformerFactory = istioinformers.NewSharedInformerFactory(istioClient, defaultResync)
	}

	if snapshotClient != nil {
		factory.snapshotInformerFactory = snapshotinformer.NewSharedInformerFactory(snapshotClient, defaultResync)
	}

	if apiextensionsClient != nil {
		factory.apiextensionsInformerFactory = apiextensionsinformers.NewSharedInformerFactory(apiextensionsClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory {
	return f.ksInformerFactory
}

func (f *informerFactories) ApplicationSharedInformerFactory() applicationinformers.SharedInformerFactory {
	return f.appInformerFactory
}

func (f *informerFactories) IstioSharedInformerFactory() istioinformers.SharedInformerFactory {
	return f.istioInformerFactory
}

func (f *informerFactories) SnapshotSharedInformerFactory() snapshotinformer.SharedInformerFactory {
	return f.snapshotInformerFactory
}

func (f *informerFactories) ApiExtensionSharedInformerFactory() apiextensionsinformers.SharedInformerFactory {
	return f.apiextensionsInformerFactory
}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}

	if f.ksInformerFactory != nil {
		f.ksInformerFactory.Start(stopCh)
	}

	if f.istioInformerFactory != nil {
		f.istioInformerFactory.Start(stopCh)
	}

	if f.appInformerFactory != nil {
		f.appInformerFactory.Start(stopCh)
	}

	if f.snapshotInformerFactory != nil {
		f.snapshotInformerFactory.Start(stopCh)
	}

	if f.apiextensionsInformerFactory != nil {
		f.apiextensionsInformerFactory.Start(stopCh)
	}
}

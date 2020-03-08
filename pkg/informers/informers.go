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
	applicationclient "github.com/kubernetes-sigs/application/pkg/client/clientset/versioned"
	applicationinformers "github.com/kubernetes-sigs/application/pkg/client/informers/externalversions"
	istioclient "istio.io/client-go/pkg/clientset/versioned"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"time"
)

const defaultResync = 600 * time.Second

type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory
	IstioSharedInformerFactory() istioinformers.SharedInformerFactory
	ApplicationSharedInformerFactory() applicationinformers.SharedInformerFactory

	// Start all the informer factories if not nil
	Start(stopCh <-chan struct{})
}

type informerFactories struct {
	informerFactory      k8sinformers.SharedInformerFactory
	ksInformerFactory    ksinformers.SharedInformerFactory
	istioInformerFactory istioinformers.SharedInformerFactory
	appInformerFactory   applicationinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, ksClient versioned.Interface, istioClient istioclient.Interface, appClient applicationclient.Interface) InformerFactory {
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

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.informerFactory != nil {
		f.informerFactory.Start(stopCh)
	}

	if f.ksInformerFactory != nil {
		f.ksInformerFactory.Start(stopCh)
	}

	if f.informerFactory != nil {
		f.istioInformerFactory.Start(stopCh)
	}

	if f.appInformerFactory != nil {
		f.appInformerFactory.Start(stopCh)
	}
}

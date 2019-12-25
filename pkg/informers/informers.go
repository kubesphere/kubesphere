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
	s2i "github.com/kubesphere/s2ioperator/pkg/client/clientset/versioned"
	s2iinformers "github.com/kubesphere/s2ioperator/pkg/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"time"
)

const defaultResync = 600 * time.Second

type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	S2iSharedInformerFactory() s2iinformers.SharedInformerFactory
	KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory
	ApplicationSharedInformerFactory() applicationinformers.SharedInformerFactory
}

type informerFactories struct {
	informerFactory    k8sinformers.SharedInformerFactory
	s2iInformerFactory s2iinformers.SharedInformerFactory
	ksInformerFactory  ksinformers.SharedInformerFactory
	appInformerFactory applicationinformers.SharedInformerFactory
}

func NewInformerFactories(client kubernetes.Interface, ksClient versioned.Interface, s2iClient s2i.Interface, appClient applicationclient.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.informerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}

	if ksClient != nil {
		factory.ksInformerFactory = ksinformers.NewSharedInformerFactory(ksClient, defaultResync)
	}

	if s2iClient != nil {
		factory.s2iInformerFactory = s2iinformers.NewSharedInformerFactory(s2iClient, defaultResync)
	}

	if appClient != nil {
		factory.appInformerFactory = applicationinformers.NewSharedInformerFactory(appClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.informerFactory
}

func (f *informerFactories) S2iSharedInformerFactory() s2iinformers.SharedInformerFactory {
	return f.s2iInformerFactory
}

func (f *informerFactories) KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory {
	return f.ksInformerFactory
}

func (f *informerFactories) ApplicationSharedInformerFactory() applicationinformers.SharedInformerFactory {
	return f.appInformerFactory
}

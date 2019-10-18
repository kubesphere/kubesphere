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
	applicationinformers "github.com/kubernetes-sigs/application/pkg/client/informers/externalversions"
	s2iinformers "github.com/kubesphere/s2ioperator/pkg/client/informers/externalversions"
	k8sinformers "k8s.io/client-go/informers"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"sync"
	"time"
)

const defaultResync = 600 * time.Second

var (
	k8sOnce            sync.Once
	s2iOnce            sync.Once
	ksOnce             sync.Once
	appOnce            sync.Once
	informerFactory    k8sinformers.SharedInformerFactory
	s2iInformerFactory s2iinformers.SharedInformerFactory
	ksInformerFactory  ksinformers.SharedInformerFactory
	appInformerFactory applicationinformers.SharedInformerFactory
)

func SharedInformerFactory() k8sinformers.SharedInformerFactory {
	k8sOnce.Do(func() {
		k8sClient := client.ClientSets().K8s().Kubernetes()
		informerFactory = k8sinformers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return informerFactory
}

func S2iSharedInformerFactory() s2iinformers.SharedInformerFactory {
	s2iOnce.Do(func() {
		k8sClient := client.ClientSets().K8s().S2i()
		s2iInformerFactory = s2iinformers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return s2iInformerFactory
}

func KsSharedInformerFactory() ksinformers.SharedInformerFactory {
	ksOnce.Do(func() {
		k8sClient := client.ClientSets().K8s().KubeSphere()
		ksInformerFactory = ksinformers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return ksInformerFactory
}

func AppSharedInformerFactory() applicationinformers.SharedInformerFactory {
	appOnce.Do(func() {
		appClient := client.ClientSets().K8s().Application()
		appInformerFactory = applicationinformers.NewSharedInformerFactory(appClient, defaultResync)
	})
	return appInformerFactory
}

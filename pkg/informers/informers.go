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
	"sync"
	"time"

	s2iInformers "github.com/kubesphere/s2ioperator/pkg/client/informers/externalversions"

	"k8s.io/client-go/informers"
	ksInformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"

	"kubesphere.io/kubesphere/pkg/simple/client/k8s"
)

const defaultResync = 600 * time.Second

var (
	k8sOnce            sync.Once
	s2iOnce            sync.Once
	ksOnce             sync.Once
	informerFactory    informers.SharedInformerFactory
	s2iInformerFactory s2iInformers.SharedInformerFactory
	ksInformerFactory  ksInformers.SharedInformerFactory
)

func SharedInformerFactory() informers.SharedInformerFactory {
	k8sOnce.Do(func() {
		k8sClient := k8s.Client()
		informerFactory = informers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return informerFactory
}

func S2iSharedInformerFactory() s2iInformers.SharedInformerFactory {
	s2iOnce.Do(func() {
		k8sClient := k8s.S2iClient()
		s2iInformerFactory = s2iInformers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return s2iInformerFactory
}

func KsSharedInformerFactory() ksInformers.SharedInformerFactory {
	ksOnce.Do(func() {
		k8sClient := k8s.KsClient()
		ksInformerFactory = ksInformers.NewSharedInformerFactory(k8sClient, defaultResync)
	})
	return ksInformerFactory
}

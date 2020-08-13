/*
Copyright 2020 KubeSphere Authors

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
	snapshotinformer "github.com/kubernetes-csi/external-snapshotter/v2/pkg/client/informers/externalversions"
	istioinformers "istio.io/client-go/pkg/informers/externalversions"
	apiextensionsinformers "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	ksfake "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	ksinformers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	appinformers "sigs.k8s.io/application/pkg/client/informers/externalversions"
	"time"
)

type nullInformerFactory struct {
	fakeK8sInformerFactory informers.SharedInformerFactory
	fakeKsInformerFactory  ksinformers.SharedInformerFactory
}

func NewNullInformerFactory() InformerFactory {
	fakeClient := fake.NewSimpleClientset()
	fakeInformerFactory := informers.NewSharedInformerFactory(fakeClient, time.Minute*10)

	fakeKsClient := ksfake.NewSimpleClientset()
	fakeKsInformerFactory := ksinformers.NewSharedInformerFactory(fakeKsClient, time.Minute*10)

	return &nullInformerFactory{
		fakeK8sInformerFactory: fakeInformerFactory,
		fakeKsInformerFactory:  fakeKsInformerFactory,
	}
}

func (n nullInformerFactory) KubernetesSharedInformerFactory() informers.SharedInformerFactory {
	return n.fakeK8sInformerFactory
}

func (n nullInformerFactory) KubeSphereSharedInformerFactory() ksinformers.SharedInformerFactory {
	return n.fakeKsInformerFactory
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

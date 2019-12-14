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

package testing

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
)

var (
	AlwaysReady      = func() bool { return true }
	ResyncPeriodFunc = func() time.Duration { return 1 * time.Second }
)

type FakeControllerBuilder struct {
	KsClient    *fake.Clientset
	KubeClient  *k8sfake.Clientset
	Kubeobjects []runtime.Object
	CRDObjects  []runtime.Object
}

func NewFakeControllerBuilder() *FakeControllerBuilder {
	return &FakeControllerBuilder{
		Kubeobjects: make([]runtime.Object, 0),
		CRDObjects:  make([]runtime.Object, 0),
	}
}

func (f *FakeControllerBuilder) NewControllerInformer() (informers.SharedInformerFactory, kubeinformers.SharedInformerFactory) {
	f.KsClient = fake.NewSimpleClientset(f.CRDObjects...)
	f.KubeClient = k8sfake.NewSimpleClientset(f.Kubeobjects...)
	i := informers.NewSharedInformerFactory(f.KsClient, ResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.KubeClient, ResyncPeriodFunc())
	return i, k8sI
}

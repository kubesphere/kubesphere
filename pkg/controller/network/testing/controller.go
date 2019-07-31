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

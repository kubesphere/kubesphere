package resources

import (
	"sync"
	"time"

	extensions "k8s.io/client-go/listers/extensions/v1beta1"

	"github.com/golang/glog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	appslisters "k8s.io/client-go/listers/apps/v1"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	cronjob "k8s.io/client-go/listers/batch/v1beta1"
	corelisters "k8s.io/client-go/listers/core/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

const defaultResync = 600 * time.Second

var (
	NamespaceLister             corelisters.NamespaceLister
	NodeLister                  corelisters.NodeLister
	DeploymentLister            appslisters.DeploymentLister
	StatefulSetLister           appslisters.StatefulSetLister
	DaemonSetLister             appslisters.DaemonSetLister
	JobLister                   batchlisters.JobLister
	CronJobLister               cronjob.CronJobLister
	ServiceLister               corelisters.ServiceLister
	IngressLister               extensions.IngressLister
	PersistentVolumeClaimLister corelisters.PersistentVolumeClaimLister
	SecretLister                corelisters.SecretLister
	ConfigMapLister             corelisters.ConfigMapLister
)

var sharedInformerFactory informers.SharedInformerFactory

var once = sync.Once{}

func SharedInformerFactory() informers.SharedInformerFactory {
	once.Do(func() {
		kubeclientset := client.NewK8sClient()
		sharedInformerFactory = informers.NewSharedInformerFactory(kubeclientset, defaultResync)
	})
	return sharedInformerFactory
}

func Sync(stopCh <-chan struct{}) {
	go func() {
		sharedInformerFactory := SharedInformerFactory()
		NamespaceLister = sharedInformerFactory.Core().V1().Namespaces().Lister()
		NodeLister = sharedInformerFactory.Core().V1().Nodes().Lister()
		DeploymentLister = sharedInformerFactory.Apps().V1().Deployments().Lister()
		StatefulSetLister = sharedInformerFactory.Apps().V1().StatefulSets().Lister()
		DaemonSetLister = sharedInformerFactory.Apps().V1().DaemonSets().Lister()
		JobLister = sharedInformerFactory.Batch().V1().Jobs().Lister()
		CronJobLister = sharedInformerFactory.Batch().V1beta1().CronJobs().Lister()
		ServiceLister = sharedInformerFactory.Core().V1().Services().Lister()
		IngressLister = sharedInformerFactory.Extensions().V1beta1().Ingresses().Lister()
		PersistentVolumeClaimLister = sharedInformerFactory.Core().V1().PersistentVolumeClaims().Lister()
		SecretLister = sharedInformerFactory.Core().V1().Secrets().Lister()
		ConfigMapLister = sharedInformerFactory.Core().V1().ConfigMaps().Lister()
		sharedInformerFactory.Start(stopCh)
		glog.Info("started resources sync")
		<-stopCh
		glog.Info("shutting resources sync")
	}()
}

func Namespaces() ([]*v1.Namespace, error) {
	return NamespaceLister.List(labels.Everything())
}

func Nodes() ([]*v1.Node, error) {
	return NodeLister.List(labels.Everything())
}

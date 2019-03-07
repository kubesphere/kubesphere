package controller

import (
	"kubesphere.io/kubesphere/pkg/controller/namespace"
	"log"
	"sync"
	"time"

	"k8s.io/client-go/informers"

	"kubesphere.io/kubesphere/pkg/client"
)

const defaultResync = 600 * time.Second

var once sync.Once

func Run(stopCh <-chan struct{}) {
	once.Do(func() {
		kubeclientset := client.K8sClient()
		informerFactory := informers.NewSharedInformerFactory(kubeclientset, defaultResync)
		namespaceController := namespace.NewNamespaceController(kubeclientset, informerFactory.Core().V1().Namespaces(), informerFactory.Rbac().V1().Roles())
		// data sync
		informerFactory.Start(stopCh)
		// start workers
		namespaceController.Start(stopCh)
		log.Println("all controller is running")
	})
}

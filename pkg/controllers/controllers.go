package controllers

import (
	"time"

	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/resources"
)

const defaultResync = 600 * time.Second

func Run(stopCh <-chan struct{}) {
	go func() {
		kubeclientset := client.NewK8sClient()
		informerFactory := resources.SharedInformerFactory()
		namespaceController := NewNamespaceController(kubeclientset, informerFactory.Core().V1().Namespaces(), informerFactory.Rbac().V1().Roles())

		// data sync
		informerFactory.Start(stopCh)
		// start workers
		namespaceController.Start(stopCh)

		<-stopCh
		glog.Info("shutting down controllers")
	}()
}

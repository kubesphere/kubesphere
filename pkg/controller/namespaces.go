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
package controller

import (
	"fmt"
	"time"

	"github.com/golang/glog"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/informers/rbac/v1"
	rbacinformers "k8s.io/client-go/informers/rbac/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const threadiness = 2

var (
	defaultRoles = []rbac.Role{
		{ObjectMeta: metaV1.ObjectMeta{Name: "admin", Annotations: map[string]string{"creator": "system"}}, Rules: []rbac.PolicyRule{{Verbs: []string{"*"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}},
		{ObjectMeta: metaV1.ObjectMeta{Name: "operator", Annotations: map[string]string{"creator": "system"}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}}, {Verbs: []string{"*"}, APIGroups: []string{"", "apps", "extensions", "batch", "kubesphere.io", "account.kubesphere.io", "autoscaling"}, Resources: []string{"*"}}}},
		{ObjectMeta: metaV1.ObjectMeta{Name: "viewer", Annotations: map[string]string{"creator": "system"}}, Rules: []rbac.PolicyRule{{Verbs: []string{"get", "list", "watch"}, APIGroups: []string{"*"}, Resources: []string{"*"}}}},
	}
)

type NamespaceController struct {
	clientset         kubernetes.Interface
	namespaceInformer coreinformers.NamespaceInformer
	roleInformer      v1.RoleInformer
	workqueue         workqueue.RateLimitingInterface
}

func NewNamespaceController(
	kubeclientset kubernetes.Interface,
	namespaceInformer coreinformers.NamespaceInformer,
	roleInformer rbacinformers.RoleInformer) *NamespaceController {

	controller := &NamespaceController{
		clientset:         kubeclientset,
		namespaceInformer: namespaceInformer,
		roleInformer:      roleInformer,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "namespaces"),
	}

	glog.Info("setting up event handlers")

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newNamespace := new.(*corev1.Namespace)
			oldNamespace := old.(*corev1.Namespace)
			if newNamespace.ResourceVersion == oldNamespace.ResourceVersion {
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	roleInformer.Lister()

	return controller
}

func (c *NamespaceController) Start(stopCh <-chan struct{}) {
	go func() {
		defer utilruntime.HandleCrash()
		defer c.workqueue.ShutDown()

		// Start the informer factories to begin populating the informer caches
		glog.Info("starting namespace controller")

		// Wait for the caches to be synced before starting workers
		glog.Info("waiting for informer caches to sync")
		if ok := cache.WaitForCacheSync(stopCh, c.namespaceInformer.Informer().HasSynced, c.roleInformer.Informer().HasSynced); !ok {
			glog.Fatalf("controller exit with error: failed to wait for caches to sync")
		}

		glog.Info("starting workers")

		for i := 0; i < threadiness; i++ {
			go wait.Until(c.runWorker, time.Second, stopCh)
		}

		glog.Info("started workers")
		<-stopCh
		glog.Info("shutting down workers")
	}()
}

func (c *NamespaceController) runWorker() {
	for c.processNextWorkItem() {
	}
}

func (c *NamespaceController) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.workqueue.Done(obj)
		var namespace string
		var ok bool

		if namespace, ok = obj.(string); !ok {
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}

		if err := c.syncHandler(namespace); err != nil {
			c.workqueue.AddRateLimited(namespace)
			return fmt.Errorf("error syncing '%s': %s, requeuing", namespace, err.Error())
		}

		c.workqueue.Forget(obj)
		glog.Infof("successfully namespace synced '%s'", namespace)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

func (c *NamespaceController) syncHandler(name string) error {

	_, err := c.namespaceInformer.Lister().Get(name)

	// Handler delete event
	if errors.IsNotFound(err) {
		return nil
	}

	// Handler update or create event
	if err := c.checkRoles(name); err != nil {
		return err
	}

	return nil
}

func (c *NamespaceController) handleObject(obj interface{}) {
	if namespace, ok := obj.(*corev1.Namespace); ok {
		c.workqueue.AddRateLimited(namespace.Name)
	}
}

func (c *NamespaceController) checkRoles(namespace string) error {
	for _, role := range defaultRoles {
		_, err := c.roleInformer.Lister().Roles(namespace).Get(role.Name)
		if errors.IsNotFound(err) {
			r := role.DeepCopy()
			r.Namespace = namespace
			_, err := c.clientset.RbacV1().Roles(namespace).Create(r)
			if err != nil && !errors.IsAlreadyExists(err) {
				return err
			}
		}
	}
	return nil
}

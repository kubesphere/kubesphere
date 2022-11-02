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

package controller

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"

	// FailedSynced is used as part of the Event 'reason' when a Foo is not synced
	FailedSynced = "FailedSync"
	// is synced successfully
	MessageResourceSynced = "Synced successfully"
)

// BaseController provides a Controller template for watching a primary resources that defined as CRD.
type BaseController struct {
	// Workers will wait informer caches to be synced
	Synced []cache.InformerSynced
	// Workqueue is a rate limited work queue.
	Workqueue  workqueue.RateLimitingInterface
	Handler    func(key string) error
	MaxRetries int
	Name       string
}

// Run will set up the event handlers for Primary resource, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *BaseController) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.Workqueue.ShutDown()

	klog.Infof("Starting controller: %s", c.Name)
	klog.Infof("Waiting for informer caches to sync for: %s", c.Name)
	if ok := cache.WaitForCacheSync(stopCh, c.Synced...); !ok {
		return fmt.Errorf("failed to wait for caches to sync for: %s", c.Name)
	}

	klog.Infof("Starting workers for: %s", c.Name)
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Infof("Started workers for: %s", c.Name)
	<-stopCh
	klog.Infof("Shutting down workers for: %s", c.Name)
	return nil
}

// Enqueue takes a primary resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than primary resource.
func (c *BaseController) Enqueue(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.Workqueue.Add(key)
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *BaseController) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the Handler.
func (c *BaseController) processNextWorkItem() bool {
	obj, shutdown := c.Workqueue.Get()

	if shutdown {
		return false
	}
	err := func(obj interface{}) error {
		defer c.Workqueue.Done(obj)
		var key string
		var ok bool

		if key, ok = obj.(string); !ok {
			c.Workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in Workqueue but got %#v in %s", obj, c.Name))
			return nil
		}
		if err := c.Handler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors,
			// when the max retries haven't reached or there is no retry times limit.
			if c.MaxRetries == 0 || c.Workqueue.NumRequeues(key) < c.MaxRetries {
				c.Workqueue.AddRateLimited(key)
				return fmt.Errorf("error syncing '%s' in %s: %s, requeuing ", key, c.Name, err.Error())
			}
			klog.V(4).Infof("Dropping %s out of the queue in %s: %s", key, c.Name, err)
			utilruntime.HandleError(err)
			return nil
		}
		c.Workqueue.Forget(obj)
		klog.Infof("Successfully Synced %s:%s in %s", "key", key, c.Name)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

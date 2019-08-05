/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package watch

import (
	"fmt"
	"time"

	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	maxRetries  = 5
	CreateEvent = "create"
	UpdateEvent = "update"
	DeleteEvent = "delete"
)

type Event struct {
	ResourceName string      `json:"resourceName"`
	EventType    string      `json:"type"`
	Namespace    string      `json:"namespace"`
	ResourceType string      `json:"resourceType"`
	Item         interface{} `json:"item"`
}

// Resource watch controller
type EventWatcher struct {
	clientset clientset.Interface
	queue     workqueue.RateLimitingInterface
	informer  cache.SharedIndexInformer
	receiver  EventReceiver
}

type EventReceiver interface {
	HandleEvent(event Event)
}

func Resource(receiver EventReceiver, informer cache.SharedIndexInformer, resourceType string) *EventWatcher {
	queue := workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	var newEvent Event
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			newEvent.Namespace, newEvent.ResourceName, err = cache.SplitMetaNamespaceKey(key)
			if err != nil {
				return
			}
			newEvent.EventType = CreateEvent
			newEvent.ResourceType = resourceType
			newEvent.Item = obj
			queue.Add(newEvent)
		},
		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(old)
			if err != nil {
				return
			}
			newEvent.Namespace, newEvent.ResourceName, err = cache.SplitMetaNamespaceKey(key)
			if err != nil {
				return
			}
			newEvent.EventType = UpdateEvent
			newEvent.ResourceType = resourceType
			newEvent.Item = new
			queue.Add(newEvent)
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				return
			}
			newEvent.Namespace, newEvent.ResourceName, err = cache.SplitMetaNamespaceKey(key)
			if err != nil {
				return
			}
			newEvent.EventType = DeleteEvent
			newEvent.ResourceType = resourceType
			newEvent.Item = obj
			queue.Add(newEvent)
		},
	})
	return &EventWatcher{
		informer: informer,
		queue:    queue,
		receiver: receiver,
	}
}

// Watch starts the watch controller
func (c *EventWatcher) Watch(stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer c.queue.ShutDown()

	if !cache.WaitForCacheSync(stopCh, c.HasSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	wait.Until(c.runWorker, time.Second, stopCh)
}

// HasSynced is required for the cache.EventWatcher interface.
func (c *EventWatcher) HasSynced() bool {
	return c.informer.HasSynced()
}

// LastSyncResourceVersion is required for the cache.EventWatcher interface.
func (c *EventWatcher) LastSyncResourceVersion() string {
	return c.informer.LastSyncResourceVersion()
}

func (c *EventWatcher) runWorker() {
	for c.processNextItem() {
		// continue looping
	}
}

func (c *EventWatcher) processNextItem() bool {
	newEvent, quit := c.queue.Get()

	if quit {
		return false
	}
	defer c.queue.Done(newEvent)
	err := c.processItem(newEvent.(Event))
	if err == nil {
		// No error, reset the rate limit counters
		c.queue.Forget(newEvent)
	} else if c.queue.NumRequeues(newEvent) < maxRetries {
		c.queue.AddRateLimited(newEvent)
	} else {
		// err != nil and too many retries
		c.queue.Forget(newEvent)
		utilruntime.HandleError(err)
	}

	return true
}

// Push resource event
func (c *EventWatcher) processItem(newEvent Event) error {
	c.receiver.HandleEvent(newEvent)
	return nil
}

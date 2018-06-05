/*
Copyright 2017 The Kubernetes Authors.

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

package resourcequota

import (
	"fmt"
	"sync"
	"time"

	"github.com/golang/glog"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/clock"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/quota"
	"k8s.io/kubernetes/pkg/quota/evaluator/core"
	"k8s.io/kubernetes/pkg/quota/generic"
)

type eventType int

func (e eventType) String() string {
	switch e {
	case addEvent:
		return "add"
	case updateEvent:
		return "update"
	case deleteEvent:
		return "delete"
	default:
		return fmt.Sprintf("unknown(%d)", int(e))
	}
}

const (
	addEvent eventType = iota
	updateEvent
	deleteEvent
)

type event struct {
	eventType eventType
	obj       interface{}
	oldObj    interface{}
	gvr       schema.GroupVersionResource
}

type QuotaMonitor struct {
	// each monitor list/watches a resource and determines if we should replenish quota
	monitors    monitors
	monitorLock sync.Mutex
	// informersStarted is closed after after all of the controllers have been initialized and are running.
	// After that it is safe to start them here, before that it is not.
	informersStarted <-chan struct{}

	// stopCh drives shutdown. When a receive from it unblocks, monitors will shut down.
	// This channel is also protected by monitorLock.
	stopCh <-chan struct{}

	// running tracks whether Run() has been called.
	// it is protected by monitorLock.
	running bool

	// monitors are the producer of the resourceChanges queue
	resourceChanges workqueue.RateLimitingInterface

	// interfaces with informers
	informerFactory InformerFactory

	// list of resources to ignore
	ignoredResources map[schema.GroupResource]struct{}

	// The period that should be used to re-sync the monitored resource
	resyncPeriod controller.ResyncPeriodFunc

	// callback to alert that a change may require quota recalculation
	replenishmentFunc ReplenishmentFunc

	// maintains list of evaluators
	registry quota.Registry
}

func NewQuotaMonitor(informersStarted <-chan struct{}, informerFactory InformerFactory, ignoredResources map[schema.GroupResource]struct{}, resyncPeriod controller.ResyncPeriodFunc, replenishmentFunc ReplenishmentFunc, registry quota.Registry) *QuotaMonitor {
	return &QuotaMonitor{
		informersStarted:  informersStarted,
		informerFactory:   informerFactory,
		ignoredResources:  ignoredResources,
		resourceChanges:   workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "resource_quota_controller_resource_changes"),
		resyncPeriod:      resyncPeriod,
		replenishmentFunc: replenishmentFunc,
		registry:          registry,
	}
}

// monitor runs a Controller with a local stop channel.
type monitor struct {
	controller cache.Controller

	// stopCh stops Controller. If stopCh is nil, the monitor is considered to be
	// not yet started.
	stopCh chan struct{}
}

// Run is intended to be called in a goroutine. Multiple calls of this is an
// error.
func (m *monitor) Run() {
	m.controller.Run(m.stopCh)
}

type monitors map[schema.GroupVersionResource]*monitor

func (qm *QuotaMonitor) controllerFor(resource schema.GroupVersionResource) (cache.Controller, error) {
	// TODO: pass this down
	clock := clock.RealClock{}
	handlers := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			// TODO: leaky abstraction!  live w/ it for now, but should pass down an update filter func.
			// we only want to queue the updates we care about though as too much noise will overwhelm queue.
			notifyUpdate := false
			switch resource.GroupResource() {
			case schema.GroupResource{Resource: "pods"}:
				oldPod := oldObj.(*v1.Pod)
				newPod := newObj.(*v1.Pod)
				notifyUpdate = core.QuotaV1Pod(oldPod, clock) && !core.QuotaV1Pod(newPod, clock)
			case schema.GroupResource{Resource: "services"}:
				oldService := oldObj.(*v1.Service)
				newService := newObj.(*v1.Service)
				notifyUpdate = core.GetQuotaServiceType(oldService) != core.GetQuotaServiceType(newService)
			}
			if notifyUpdate {
				event := &event{
					eventType: updateEvent,
					obj:       newObj,
					oldObj:    oldObj,
					gvr:       resource,
				}
				qm.resourceChanges.Add(event)
			}
		},
		DeleteFunc: func(obj interface{}) {
			// delta fifo may wrap the object in a cache.DeletedFinalStateUnknown, unwrap it
			if deletedFinalStateUnknown, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				obj = deletedFinalStateUnknown.Obj
			}
			event := &event{
				eventType: deleteEvent,
				obj:       obj,
				gvr:       resource,
			}
			qm.resourceChanges.Add(event)
		},
	}
	shared, err := qm.informerFactory.ForResource(resource)
	if err == nil {
		glog.V(4).Infof("QuotaMonitor using a shared informer for resource %q", resource.String())
		shared.Informer().AddEventHandlerWithResyncPeriod(handlers, qm.resyncPeriod())
		return shared.Informer().GetController(), nil
	}
	glog.V(4).Infof("QuotaMonitor unable to use a shared informer for resource %q: %v", resource.String(), err)

	// TODO: if we can share storage with garbage collector, it may make sense to support other resources
	// until that time, aggregated api servers will have to run their own controller to reconcile their own quota.
	return nil, fmt.Errorf("unable to monitor quota for resource %q", resource.String())
}

// SyncMonitors rebuilds the monitor set according to the supplied resources,
// creating or deleting monitors as necessary. It will return any error
// encountered, but will make an attempt to create a monitor for each resource
// instead of immediately exiting on an error. It may be called before or after
// Run. Monitors are NOT started as part of the sync. To ensure all existing
// monitors are started, call StartMonitors.
func (qm *QuotaMonitor) SyncMonitors(resources map[schema.GroupVersionResource]struct{}) error {
	qm.monitorLock.Lock()
	defer qm.monitorLock.Unlock()

	toRemove := qm.monitors
	if toRemove == nil {
		toRemove = monitors{}
	}
	current := monitors{}
	errs := []error{}
	kept := 0
	added := 0
	for resource := range resources {
		if _, ok := qm.ignoredResources[resource.GroupResource()]; ok {
			continue
		}
		if m, ok := toRemove[resource]; ok {
			current[resource] = m
			delete(toRemove, resource)
			kept++
			continue
		}
		c, err := qm.controllerFor(resource)
		if err != nil {
			errs = append(errs, fmt.Errorf("couldn't start monitor for resource %q: %v", resource, err))
			continue
		}

		// check if we need to create an evaluator for this resource (if none previously registered)
		evaluator := qm.registry.Get(resource.GroupResource())
		if evaluator == nil {
			listerFunc := generic.ListerFuncForResourceFunc(qm.informerFactory.ForResource)
			listResourceFunc := generic.ListResourceUsingListerFunc(listerFunc, resource)
			evaluator = generic.NewObjectCountEvaluator(false, resource.GroupResource(), listResourceFunc, "")
			qm.registry.Add(evaluator)
			glog.Infof("QuotaMonitor created object count evaluator for %s", resource.GroupResource())
		}

		// track the monitor
		current[resource] = &monitor{controller: c}
		added++
	}
	qm.monitors = current

	for _, monitor := range toRemove {
		if monitor.stopCh != nil {
			close(monitor.stopCh)
		}
	}

	glog.V(4).Infof("quota synced monitors; added %d, kept %d, removed %d", added, kept, len(toRemove))
	// NewAggregate returns nil if errs is 0-length
	return utilerrors.NewAggregate(errs)
}

// StartMonitors ensures the current set of monitors are running. Any newly
// started monitors will also cause shared informers to be started.
//
// If called before Run, StartMonitors does nothing (as there is no stop channel
// to support monitor/informer execution).
func (qm *QuotaMonitor) StartMonitors() {
	qm.monitorLock.Lock()
	defer qm.monitorLock.Unlock()

	if !qm.running {
		return
	}

	// we're waiting until after the informer start that happens once all the controllers are initialized.  This ensures
	// that they don't get unexpected events on their work queues.
	<-qm.informersStarted

	monitors := qm.monitors
	started := 0
	for _, monitor := range monitors {
		if monitor.stopCh == nil {
			monitor.stopCh = make(chan struct{})
			qm.informerFactory.Start(qm.stopCh)
			go monitor.Run()
			started++
		}
	}
	glog.V(4).Infof("QuotaMonitor started %d new monitors, %d currently running", started, len(monitors))
}

// IsSynced returns true if any monitors exist AND all those monitors'
// controllers HasSynced functions return true. This means IsSynced could return
// true at one time, and then later return false if all monitors were
// reconstructed.
func (qm *QuotaMonitor) IsSynced() bool {
	qm.monitorLock.Lock()
	defer qm.monitorLock.Unlock()

	if len(qm.monitors) == 0 {
		return false
	}

	for _, monitor := range qm.monitors {
		if !monitor.controller.HasSynced() {
			return false
		}
	}
	return true
}

// Run sets the stop channel and starts monitor execution until stopCh is
// closed. Any running monitors will be stopped before Run returns.
func (qm *QuotaMonitor) Run(stopCh <-chan struct{}) {
	glog.Infof("QuotaMonitor running")
	defer glog.Infof("QuotaMonitor stopping")

	// Set up the stop channel.
	qm.monitorLock.Lock()
	qm.stopCh = stopCh
	qm.running = true
	qm.monitorLock.Unlock()

	// Start monitors and begin change processing until the stop channel is
	// closed.
	qm.StartMonitors()
	wait.Until(qm.runProcessResourceChanges, 1*time.Second, stopCh)

	// Stop any running monitors.
	qm.monitorLock.Lock()
	defer qm.monitorLock.Unlock()
	monitors := qm.monitors
	stopped := 0
	for _, monitor := range monitors {
		if monitor.stopCh != nil {
			stopped++
			close(monitor.stopCh)
		}
	}
	glog.Infof("QuotaMonitor stopped %d of %d monitors", stopped, len(monitors))
}

func (qm *QuotaMonitor) runProcessResourceChanges() {
	for qm.processResourceChanges() {
	}
}

// Dequeueing an event from resourceChanges to process
func (qm *QuotaMonitor) processResourceChanges() bool {
	item, quit := qm.resourceChanges.Get()
	if quit {
		return false
	}
	defer qm.resourceChanges.Done(item)
	event, ok := item.(*event)
	if !ok {
		utilruntime.HandleError(fmt.Errorf("expect a *event, got %v", item))
		return true
	}
	obj := event.obj
	accessor, err := meta.Accessor(obj)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("cannot access obj: %v", err))
		return true
	}
	glog.V(4).Infof("QuotaMonitor process object: %s, namespace %s, name %s, uid %s, event type %v", event.gvr.String(), accessor.GetNamespace(), accessor.GetName(), string(accessor.GetUID()), event.eventType)
	qm.replenishmentFunc(event.gvr.GroupResource(), accessor.GetNamespace())
	return true
}

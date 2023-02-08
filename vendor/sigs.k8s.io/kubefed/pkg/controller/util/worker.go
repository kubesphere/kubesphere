/*
Copyright 2018 The Kubernetes Authors.

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

package util

import (
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/client-go/util/workqueue"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/kubefed/pkg/metrics"
)

type ReconcileFunc func(qualifiedName QualifiedName) ReconciliationStatus

type ReconcileWorker interface {
	Enqueue(qualifiedName QualifiedName)
	EnqueueForClusterSync(qualifiedName QualifiedName)
	EnqueueForError(qualifiedName QualifiedName)
	EnqueueForRetry(qualifiedName QualifiedName)
	EnqueueObject(obj runtimeclient.Object)
	EnqueueWithDelay(qualifiedName QualifiedName, delay time.Duration)
	Run(stopChan <-chan struct{})
	SetDelay(retryDelay, clusterSyncDelay time.Duration)
}

type WorkerOptions struct {
	WorkerTiming

	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.
	MaxConcurrentReconciles int
}

type WorkerTiming struct {
	Interval         time.Duration
	RetryDelay       time.Duration
	ClusterSyncDelay time.Duration
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
}

type asyncWorker struct {
	name string

	reconcile ReconcileFunc

	timing WorkerTiming

	maxConcurrentReconciles int

	// For triggering reconciliation of a single resource. This is
	// used when there is an add/update/delete operation on a resource
	// in either the API of the cluster hosting KubeFed or in the API
	// of a member cluster.
	deliverer *DelayingDeliverer

	// Work queue allowing parallel processing of resources
	queue workqueue.Interface

	// Backoff manager
	backoff *flowcontrol.Backoff
}

func NewReconcileWorker(name string, reconcile ReconcileFunc, options WorkerOptions) ReconcileWorker {
	if options.Interval == 0 {
		options.Interval = time.Second * 1
	}
	if options.RetryDelay == 0 {
		options.RetryDelay = time.Second * 10
	}
	if options.InitialBackoff == 0 {
		options.InitialBackoff = time.Second * 5
	}
	if options.MaxBackoff == 0 {
		options.MaxBackoff = time.Minute
	}
	if options.MaxConcurrentReconciles == 0 {
		options.MaxConcurrentReconciles = 1
	}
	return &asyncWorker{
		name:                    name,
		reconcile:               reconcile,
		timing:                  options.WorkerTiming,
		maxConcurrentReconciles: options.MaxConcurrentReconciles,
		deliverer:               NewDelayingDeliverer(),
		queue:                   workqueue.NewNamed(name),
		backoff:                 flowcontrol.NewBackOff(options.InitialBackoff, options.MaxBackoff),
	}
}

func (w *asyncWorker) Enqueue(qualifiedName QualifiedName) {
	w.deliver(qualifiedName, 0, false)
}

func (w *asyncWorker) EnqueueForError(qualifiedName QualifiedName) {
	w.deliver(qualifiedName, 0, true)
}

func (w *asyncWorker) EnqueueForRetry(qualifiedName QualifiedName) {
	w.deliver(qualifiedName, w.timing.RetryDelay, false)
}

func (w *asyncWorker) EnqueueForClusterSync(qualifiedName QualifiedName) {
	w.deliver(qualifiedName, w.timing.ClusterSyncDelay, false)
}

func (w *asyncWorker) EnqueueObject(obj runtimeclient.Object) {
	qualifiedName := NewQualifiedName(obj)
	w.Enqueue(qualifiedName)
}

func (w *asyncWorker) EnqueueWithDelay(qualifiedName QualifiedName, delay time.Duration) {
	w.deliver(qualifiedName, delay, false)
}

func (w *asyncWorker) Run(stopChan <-chan struct{}) {
	w.initMetrics()

	StartBackoffGC(w.backoff, stopChan)
	w.deliverer.StartWithHandler(func(item *DelayingDelivererItem) {
		qualifiedName, ok := item.Value.(*QualifiedName)
		if ok {
			w.queue.Add(*qualifiedName)
		}
	})

	for i := 0; i < w.maxConcurrentReconciles; i++ {
		go wait.Until(w.worker, w.timing.Interval, stopChan)
	}

	// Ensure all goroutines are cleaned up when the stop channel closes
	go func() {
		<-stopChan
		w.queue.ShutDown()
		w.deliverer.Stop()
	}()
}

func (w *asyncWorker) SetDelay(retryDelay, clusterSyncDelay time.Duration) {
	w.timing.RetryDelay = retryDelay
	w.timing.ClusterSyncDelay = clusterSyncDelay
}

// deliver adds backoff to delay if this delivery is related to some
// failure. Resets backoff if there was no failure.
func (w *asyncWorker) deliver(qualifiedName QualifiedName, delay time.Duration, failed bool) {
	key := qualifiedName.String()
	if failed {
		w.backoff.Next(key, time.Now())
		delay += w.backoff.Get(key)
	} else {
		w.backoff.Reset(key)
	}
	w.deliverer.DeliverAfter(key, &qualifiedName, delay)
}

func (w *asyncWorker) worker() {
	for w.reconcileOnce() {
	}
}

func (w *asyncWorker) reconcileOnce() bool {
	obj, quit := w.queue.Get()
	if quit {
		return false
	}
	defer w.queue.Done(obj)

	qualifiedName, ok := obj.(QualifiedName)
	if !ok {
		return true
	}

	metrics.ControllerRuntimeActiveWorkers.WithLabelValues(w.name).Add(1)
	defer metrics.ControllerRuntimeActiveWorkers.WithLabelValues(w.name).Add(-1)
	defer metrics.UpdateControllerRuntimeReconcileTimeFromStart(w.name, time.Now())

	status := w.reconcile(qualifiedName)
	switch status {
	case StatusAllOK:
		metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelSuccess).Inc()
	case StatusError:
		w.EnqueueForError(qualifiedName)
		metrics.ControllerRuntimeReconcileErrors.WithLabelValues(w.name).Inc()
		metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelError).Inc()
	case StatusNeedsRecheck:
		w.EnqueueForRetry(qualifiedName)
		metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelNeedsRecheck).Inc()
	case StatusNotSynced:
		w.EnqueueForClusterSync(qualifiedName)
		metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelNotSynced).Inc()
	}
	return true
}

const (
	labelSuccess      = "success"
	labelError        = "error"
	labelNeedsRecheck = "needs_recheck"
	labelNotSynced    = "not_synced"
)

func (w *asyncWorker) initMetrics() {
	metrics.ControllerRuntimeActiveWorkers.WithLabelValues(w.name).Set(0)
	metrics.ControllerRuntimeReconcileErrors.WithLabelValues(w.name).Add(0)
	metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelSuccess).Add(0)
	metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelError).Add(0)
	metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelNeedsRecheck).Add(0)
	metrics.ControllerRuntimeReconcileTotal.WithLabelValues(w.name, labelNotSynced).Add(0)
	metrics.ControllerRuntimeWorkerCount.WithLabelValues(w.name).Set(float64(w.maxConcurrentReconciles))
}

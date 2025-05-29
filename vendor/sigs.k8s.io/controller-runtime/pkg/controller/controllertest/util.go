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

package controllertest

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

var _ cache.SharedIndexInformer = &FakeInformer{}

// FakeInformer provides fake Informer functionality for testing.
type FakeInformer struct {
	// Synced is returned by the HasSynced functions to implement the Informer interface
	Synced bool

	// RunCount is incremented each time RunInformersAndControllers is called
	RunCount int

	handlers []cache.ResourceEventHandler
}

// AddIndexers does nothing.  TODO(community): Implement this.
func (f *FakeInformer) AddIndexers(indexers cache.Indexers) error {
	return nil
}

// GetIndexer does nothing.  TODO(community): Implement this.
func (f *FakeInformer) GetIndexer() cache.Indexer {
	return nil
}

// Informer returns the fake Informer.
func (f *FakeInformer) Informer() cache.SharedIndexInformer {
	return f
}

// HasSynced implements the Informer interface.  Returns f.Synced.
func (f *FakeInformer) HasSynced() bool {
	return f.Synced
}

// AddEventHandler implements the Informer interface. Adds an EventHandler to the fake Informers. TODO(community): Implement Registration.
func (f *FakeInformer) AddEventHandler(handler cache.ResourceEventHandler) (cache.ResourceEventHandlerRegistration, error) {
	f.handlers = append(f.handlers, handler)
	return nil, nil
}

// AddEventHandlerWithResyncPeriod implements the Informer interface. Adds an EventHandler to the fake Informers (ignores resyncPeriod). TODO(community): Implement Registration.
func (f *FakeInformer) AddEventHandlerWithResyncPeriod(handler cache.ResourceEventHandler, _ time.Duration) (cache.ResourceEventHandlerRegistration, error) {
	f.handlers = append(f.handlers, handler)
	return nil, nil
}

// AddEventHandlerWithOptions implements the Informer interface. Adds an EventHandler to the fake Informers (ignores options). TODO(community): Implement Registration.
func (f *FakeInformer) AddEventHandlerWithOptions(handler cache.ResourceEventHandler, _ cache.HandlerOptions) (cache.ResourceEventHandlerRegistration, error) {
	f.handlers = append(f.handlers, handler)
	return nil, nil
}

// Run implements the Informer interface.  Increments f.RunCount.
func (f *FakeInformer) Run(<-chan struct{}) {
	f.RunCount++
}

func (f *FakeInformer) RunWithContext(_ context.Context) {
	f.RunCount++
}

// Add fakes an Add event for obj.
func (f *FakeInformer) Add(obj metav1.Object) {
	for _, h := range f.handlers {
		h.OnAdd(obj, false)
	}
}

// Update fakes an Update event for obj.
func (f *FakeInformer) Update(oldObj, newObj metav1.Object) {
	for _, h := range f.handlers {
		h.OnUpdate(oldObj, newObj)
	}
}

// Delete fakes an Delete event for obj.
func (f *FakeInformer) Delete(obj metav1.Object) {
	for _, h := range f.handlers {
		h.OnDelete(obj)
	}
}

// RemoveEventHandler does nothing.  TODO(community): Implement this.
func (f *FakeInformer) RemoveEventHandler(handle cache.ResourceEventHandlerRegistration) error {
	return nil
}

// GetStore does nothing.  TODO(community): Implement this.
func (f *FakeInformer) GetStore() cache.Store {
	return nil
}

// GetController does nothing.  TODO(community): Implement this.
func (f *FakeInformer) GetController() cache.Controller {
	return nil
}

// LastSyncResourceVersion does nothing.  TODO(community): Implement this.
func (f *FakeInformer) LastSyncResourceVersion() string {
	return ""
}

// SetWatchErrorHandler does nothing.  TODO(community): Implement this.
func (f *FakeInformer) SetWatchErrorHandler(cache.WatchErrorHandler) error {
	return nil
}

// SetWatchErrorHandlerWithContext does nothing.  TODO(community): Implement this.
func (f *FakeInformer) SetWatchErrorHandlerWithContext(cache.WatchErrorHandlerWithContext) error {
	return nil
}

// SetTransform does nothing.  TODO(community): Implement this.
func (f *FakeInformer) SetTransform(t cache.TransformFunc) error {
	return nil
}

// IsStopped does nothing.  TODO(community): Implement this.
func (f *FakeInformer) IsStopped() bool {
	return false
}

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

package informertest

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	toolscache "k8s.io/client-go/tools/cache"

	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllertest"
)

var _ cache.Cache = &FakeInformers{}

// FakeInformers is a fake implementation of Informers.
type FakeInformers struct {
	InformersByGVK map[schema.GroupVersionKind]toolscache.SharedIndexInformer
	Scheme         *runtime.Scheme
	Error          error
	Synced         *bool
}

// GetInformerForKind implements Informers.
func (c *FakeInformers) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind, opts ...cache.InformerGetOption) (cache.Informer, error) {
	if c.Scheme == nil {
		c.Scheme = scheme.Scheme
	}
	obj, err := c.Scheme.New(gvk)
	if err != nil {
		return nil, err
	}
	return c.informerFor(gvk, obj)
}

// FakeInformerForKind implements Informers.
func (c *FakeInformers) FakeInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (*controllertest.FakeInformer, error) {
	i, err := c.GetInformerForKind(ctx, gvk)
	if err != nil {
		return nil, err
	}
	return i.(*controllertest.FakeInformer), nil
}

// GetInformer implements Informers.
func (c *FakeInformers) GetInformer(ctx context.Context, obj client.Object, opts ...cache.InformerGetOption) (cache.Informer, error) {
	if c.Scheme == nil {
		c.Scheme = scheme.Scheme
	}
	gvks, _, err := c.Scheme.ObjectKinds(obj)
	if err != nil {
		return nil, err
	}
	gvk := gvks[0]
	return c.informerFor(gvk, obj)
}

// RemoveInformer implements Informers.
func (c *FakeInformers) RemoveInformer(ctx context.Context, obj client.Object) error {
	if c.Scheme == nil {
		c.Scheme = scheme.Scheme
	}
	gvks, _, err := c.Scheme.ObjectKinds(obj)
	if err != nil {
		return err
	}
	gvk := gvks[0]
	delete(c.InformersByGVK, gvk)
	return nil
}

// WaitForCacheSync implements Informers.
func (c *FakeInformers) WaitForCacheSync(ctx context.Context) bool {
	if c.Synced == nil {
		return true
	}
	return *c.Synced
}

// FakeInformerFor implements Informers.
func (c *FakeInformers) FakeInformerFor(ctx context.Context, obj client.Object) (*controllertest.FakeInformer, error) {
	i, err := c.GetInformer(ctx, obj)
	if err != nil {
		return nil, err
	}
	return i.(*controllertest.FakeInformer), nil
}

func (c *FakeInformers) informerFor(gvk schema.GroupVersionKind, _ runtime.Object) (toolscache.SharedIndexInformer, error) {
	if c.Error != nil {
		return nil, c.Error
	}
	if c.InformersByGVK == nil {
		c.InformersByGVK = map[schema.GroupVersionKind]toolscache.SharedIndexInformer{}
	}
	informer, ok := c.InformersByGVK[gvk]
	if ok {
		return informer, nil
	}

	c.InformersByGVK[gvk] = &controllertest.FakeInformer{}
	return c.InformersByGVK[gvk], nil
}

// Start implements Informers.
func (c *FakeInformers) Start(ctx context.Context) error {
	return c.Error
}

// IndexField implements Cache.
func (c *FakeInformers) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

// Get implements Cache.
func (c *FakeInformers) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return nil
}

// List implements Cache.
func (c *FakeInformers) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return nil
}

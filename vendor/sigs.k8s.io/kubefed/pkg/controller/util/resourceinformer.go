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
	"context"

	"github.com/pkg/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	pkgruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// NewResourceInformer returns an unfiltered informer.
func NewResourceInformer(client ResourceClient, namespace string, apiResource *metav1.APIResource, triggerFunc func(runtimeclient.Object)) (cache.Store, cache.Controller) {
	return newResourceInformer(client, namespace, apiResource, triggerFunc, "")
}

// NewManagedResourceInformer returns an informer limited to resources
// managed by KubeFed as indicated by labeling.
func NewManagedResourceInformer(client ResourceClient, namespace string, apiResource *metav1.APIResource, triggerFunc func(runtimeclient.Object)) (cache.Store, cache.Controller) {
	labelSelector := labels.Set(map[string]string{ManagedByKubeFedLabelKey: ManagedByKubeFedLabelValue}).AsSelector().String()
	return newResourceInformer(client, namespace, apiResource, triggerFunc, labelSelector)
}

func newResourceInformer(client ResourceClient, namespace string, apiResource *metav1.APIResource, triggerFunc func(runtimeclient.Object), labelSelector string) (cache.Store, cache.Controller) {
	obj := &unstructured.Unstructured{}

	if apiResource != nil {
		gvk := schema.GroupVersionKind{Group: apiResource.Group, Version: apiResource.Version, Kind: apiResource.Kind}
		obj.SetGroupVersionKind(gvk)
	}
	return cache.NewInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (pkgruntime.Object, error) {
				options.LabelSelector = labelSelector
				return client.Resources(namespace).List(context.Background(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = labelSelector
				return client.Resources(namespace).Watch(context.Background(), options)
			},
		},
		obj, // use an unstructured type with apiVersion / kind populated for informer logging purposes
		NoResyncPeriod,
		NewTriggerOnAllChanges(triggerFunc),
	)
}

func ObjFromCache(store cache.Store, kind, key string) (*unstructured.Unstructured, error) {
	obj, err := rawObjFromCache(store, kind, key)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return nil, nil
	}
	return obj.(*unstructured.Unstructured), nil
}

func rawObjFromCache(store cache.Store, kind, key string) (runtimeclient.Object, error) {
	cachedObj, exist, err := store.GetByKey(key)
	if err != nil {
		wrappedErr := errors.Wrapf(err, "Failed to query %s store for %q", kind, key)
		runtime.HandleError(wrappedErr)
		return nil, err
	}
	if !exist {
		return nil, nil
	}
	return cachedObj.(runtimeclient.Object).DeepCopyObject().(runtimeclient.Object), nil
}

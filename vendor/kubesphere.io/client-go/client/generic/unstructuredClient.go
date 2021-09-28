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

package generic

import (
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/client-go/client"
)

// client is a client.Client that reads and writes directly from/to an API server.  It lazily initializes
// new clients at the time they are used, and caches the client.
type unstructuredClient struct {
	cache      client.ClientCache
	paramCodec runtime.ParameterCodec
}

// Create implements client.Client
func (uc *unstructuredClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	createOpts := &client.CreateOptions{}
	createOpts.ApplyOptions(opts)
	result := o.Post().
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Body(obj).
		VersionedParams(createOpts.AsCreateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)
	return result
}

// Update implements client.Client
func (uc *unstructuredClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := client.UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	result := o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		Body(obj).
		VersionedParams(updateOpts.AsUpdateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)
	return result
}

// Delete implements client.Client
func (uc *unstructuredClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	deleteOpts := client.DeleteOptions{}
	deleteOpts.ApplyOptions(opts)
	return o.Delete().
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		Body(deleteOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}

// DeleteAllOf implements client.Client
func (uc *unstructuredClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	deleteAllOfOpts := client.DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)
	return o.Delete().
		NamespaceIfScoped(deleteAllOfOpts.ListOptions.Namespace, o.IsNamespaced()).
		Resource(o.Resource()).
		VersionedParams(deleteAllOfOpts.AsListOptions(), uc.paramCodec).
		Body(deleteAllOfOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}

// Patch implements client.Client
func (uc *unstructuredClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &client.PatchOptions{}
	return o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), uc.paramCodec).
		Body(data).
		Do(ctx).
		Into(obj)
}

// Get implements client.Client
func (uc *unstructuredClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	r, err := uc.cache.GetResource(obj)
	if err != nil {
		return err
	}

	result := r.Get().
		NamespaceIfScoped(key.Namespace, r.IsNamespaced()).
		Resource(r.Resource()).
		Name(key.Name).
		Do(ctx).
		Into(obj)

	u.SetGroupVersionKind(gvk)

	return result
}

// List implements client.Client
func (uc *unstructuredClient) List(ctx context.Context, obj runtime.Object, opts ...client.ListOption) error {
	u, ok := obj.(*unstructured.UnstructuredList)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()
	if strings.HasSuffix(gvk.Kind, "List") {
		gvk.Kind = gvk.Kind[:len(gvk.Kind)-4]
	}

	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	r, err := uc.cache.GetResource(obj)
	if err != nil {
		return err
	}

	return r.Get().
		NamespaceIfScoped(listOpts.Namespace, r.IsNamespaced()).
		Resource(r.Resource()).
		VersionedParams(listOpts.AsListOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)
}

func (uc *unstructuredClient) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	_, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(obj).
		VersionedParams((&client.UpdateOptions{}).ApplyOptions(opts).AsUpdateOptions(), uc.paramCodec).
		Do(ctx).
		Into(obj)
}

func (uc *unstructuredClient) PatchStatus(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unstructured client did not understand object: %T", obj)
	}

	gvk := u.GroupVersionKind()

	o, err := uc.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &client.PatchOptions{}
	result := o.Patch(patch.Type()).
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(data).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), uc.paramCodec).
		Do(ctx).
		Into(u)

	u.SetGroupVersionKind(gvk)
	return result
}

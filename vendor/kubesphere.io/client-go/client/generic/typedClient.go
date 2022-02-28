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

	"k8s.io/apimachinery/pkg/runtime"
	"kubesphere.io/client-go/client"
)

// client is a client.Client that reads and writes directly from/to an API server.  It lazily initializes
// new clients at the time they are used, and caches the client.
type typedClient struct {
	cache      client.ClientCache
	paramCodec runtime.ParameterCodec
}

// Create implements client.Client
func (c *typedClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	createOpts := &client.CreateOptions{}
	createOpts.ApplyOptions(opts)

	req := o.Post().
		Body(obj).
		VersionedParams(createOpts.AsCreateOptions(), c.paramCodec)

	// Overwrites GVK based URL when has URLOption
	if createOpts.URLOption != nil {
		absPath := createOpts.URLOption.URL()
		if createOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", createOpts.Workspace.Name)
		}
		req = req.AbsPath(absPath...)
	}
	if createOpts.URLOption == nil || createOpts.URLOption.AbsPath == "" {
		req = req.NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
			Resource(o.Resource())
	}

	return req.
		Do(ctx).
		Into(obj)
}

// Update implements client.Client
func (c *typedClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	updateOpts := &client.UpdateOptions{}
	updateOpts.ApplyOptions(opts)

	req := o.Put()

	// Overwrites GVK based URL when has URLOption
	if updateOpts.URLOption != nil {
		absPath := updateOpts.URLOption.URL()
		if updateOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", updateOpts.Workspace.Name)
		}
		req = req.AbsPath(absPath...)
	}
	if updateOpts.URLOption == nil || updateOpts.URLOption.AbsPath == "" {
		req = req.NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
			Resource(o.Resource()).
			Name(o.GetName())
	}

	return req.Body(obj).
		VersionedParams(updateOpts.AsUpdateOptions(), c.paramCodec).
		Do(ctx).
		Into(obj)
}

// Delete implements client.Client
func (c *typedClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	deleteOpts := client.DeleteOptions{}
	deleteOpts.ApplyOptions(opts)

	req := o.Delete()

	// Overwrites GVK based URL when has URLOption
	if deleteOpts.URLOption != nil {
		absPath := deleteOpts.URLOption.URL()
		if deleteOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", deleteOpts.Workspace.Name)
		}
		req = req.AbsPath(absPath...)
	}
	if deleteOpts.URLOption == nil || deleteOpts.URLOption.AbsPath == "" {
		req = req.NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
			Resource(o.Resource()).
			Name(o.GetName())
	}

	return req.Body(deleteOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}

// Patch implements client.Client
func (c *typedClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	data, err := patch.Data(obj)
	if err != nil {
		return err
	}

	patchOpts := &client.PatchOptions{}

	req := o.Patch(patch.Type())
	// Overwrites GVK based URL when has URLOption
	if patchOpts.URLOption != nil {
		absPath := patchOpts.URLOption.URL()
		if patchOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", patchOpts.Workspace.Name)
		}
		req = req.AbsPath(absPath...)
	}
	if patchOpts.URLOption == nil || patchOpts.URLOption.AbsPath == "" {
		req.NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
			Resource(o.Resource()).
			Name(o.GetName())
	}

	return req.VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), c.paramCodec).
		Body(data).
		Do(ctx).
		Into(obj)
}

// Get implements client.Client
func (c *typedClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
	r, err := c.cache.GetResource(obj)
	if err != nil {
		return err
	}

	getOpts := client.GetOptions{}
	getOpts.ApplyOptions(opts)

	req := r.Get()
	// Overwrites GVK based URL when has URLOption
	if getOpts.URLOption != nil {
		absPath := getOpts.URLOption.URL()
		if getOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", getOpts.Workspace.Name)
		}
	}
	if getOpts.URLOption == nil || getOpts.URLOption.AbsPath == "" {
		req = req.NamespaceIfScoped(key.Namespace, r.IsNamespaced()).
			Resource(r.Resource()).
			Name(key.Name)
	}

	return req.
		Do(ctx).
		Into(obj)
}

// List implements client.Client
func (c *typedClient) List(ctx context.Context, obj runtime.Object, opts ...client.ListOption) error {
	r, err := c.cache.GetResource(obj)
	if err != nil {
		return err
	}
	listOpts := client.ListOptions{}
	listOpts.ApplyOptions(opts)

	req := r.Get()

	// Overwrites GVK based URL when has URLOption
	if listOpts.URLOption != nil {
		absPath := listOpts.URLOption.URL()
		if listOpts.Workspace != nil {
			absPath = append(absPath, "workspaces", listOpts.Workspace.Name)
		}
	}
	if listOpts.URLOption == nil || listOpts.URLOption.AbsPath == "" {
		req = req.NamespaceIfScoped(listOpts.Namespace, r.IsNamespaced()).
			Resource(r.Resource()).
			VersionedParams(listOpts.AsListOptions(), c.paramCodec)
	}

	return req.
		Do(ctx).
		Into(obj)
}

// UpdateStatus used by StatusWriter to write status.
func (c *typedClient) UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}
	// It will be nice to receive an error saying the object doesn't implement
	// status subresource and check CRD definition
	return o.Put().
		NamespaceIfScoped(o.GetNamespace(), o.IsNamespaced()).
		Resource(o.Resource()).
		Name(o.GetName()).
		SubResource("status").
		Body(obj).
		VersionedParams((&client.UpdateOptions{}).ApplyOptions(opts).AsUpdateOptions(), c.paramCodec).
		Do(ctx).
		Into(obj)
}

// PatchStatus used by StatusWriter to write status.
func (c *typedClient) PatchStatus(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	o, err := c.cache.GetObjMeta(obj)
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
		SubResource("status").
		Body(data).
		VersionedParams(patchOpts.ApplyOptions(opts).AsPatchOptions(), c.paramCodec).
		Do(ctx).
		Into(obj)
}

// DeleteAllOf implements client.Client
func (c *typedClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	o, err := c.cache.GetObjMeta(obj)
	if err != nil {
		return err
	}

	deleteAllOfOpts := client.DeleteAllOfOptions{}
	deleteAllOfOpts.ApplyOptions(opts)

	return o.Delete().
		NamespaceIfScoped(deleteAllOfOpts.ListOptions.Namespace, o.IsNamespaced()).
		Resource(o.Resource()).
		VersionedParams(deleteAllOfOpts.AsListOptions(), c.paramCodec).
		Body(deleteAllOfOpts.AsDeleteOptions()).
		Do(ctx).
		Error()
}

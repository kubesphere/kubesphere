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

package resource

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"kubesphere.io/client-go/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

// client is a client.Reader that reads directly from an KubeSphere API server.  It lazily initializes
// new clients at the time they are used, and caches the client.
// TODO(Roland): Paging and sorting parameter should be supported.
type typedClient struct {
	cache      client.ClientCache
	paramCodec runtime.ParameterCodec
}

func New(config *rest.Config, options client.Options) (client.Reader, error) {
	if config == nil {
		return nil, fmt.Errorf("must provide non-nil rest.Config to client.New")
	}

	// Init a scheme if none provided
	if options.Scheme == nil {
		options.Scheme = scheme.Scheme
	}

	// Init a Mapper if none provided
	if options.Mapper == nil {
		var err error
		options.Mapper, err = apiutil.NewDynamicRESTMapper(config)
		if err != nil {
			return nil, err
		}
	}

	clientcache := client.NewClientCache(config, options)

	typedClient := &typedClient{
		cache:      clientcache,
		paramCodec: runtime.NewParameterCodec(options.Scheme),
	}
	return typedClient, nil
}

// Get implements client.Client
func (c *typedClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object, opts ...client.GetOption) error {
	r, err := c.cache.GetResource(obj)
	if err != nil {
		return err
	}

	// {ksapi}/namespaces/{namespace}/{resources}/{name}
	// {ksapi}/{resources}/{name}

	return r.Get().
		AbsPath("kapis", "resources.kubesphere.io", "v1alpha3").
		NamespaceIfScoped(key.Namespace, r.IsNamespaced()).
		Resource(r.Resource()).
		Name(key.Name).
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

	// {ksapi}/namespaces/{namespace}/{resources}
	// {ksapi}/{resources}

	return r.Get().
		AbsPath("kapis", "resources.kubesphere.io", "v1alpha3").
		NamespaceIfScoped(listOpts.Namespace, r.IsNamespaced()).
		Resource(r.Resource()).
		VersionedParams(listOpts.AsListOptions(), c.paramCodec).
		Do(ctx).
		Into(obj)
}

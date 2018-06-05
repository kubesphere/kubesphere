/*
Copyright 2016 The Kubernetes Authors.

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

// Package dynamic provides a client interface to arbitrary Kubernetes
// APIs that exposes common high level operations and exposes common
// metadata.
package dynamic

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	restclient "k8s.io/client-go/rest"
)

// Interface is a Kubernetes client that allows you to access metadata
// and manipulate metadata of a Kubernetes API group.
type Interface interface {
	// Resource returns an API interface to the specified resource for this client's
	// group and version.  If resource is not a namespaced resource, then namespace
	// is ignored.  The ResourceInterface inherits the parameter codec of this client.
	Resource(resource *metav1.APIResource, namespace string) ResourceInterface
}

// ResourceInterface is an API interface to a specific resource under a
// dynamic client.
type ResourceInterface interface {
	// List returns a list of objects for this resource.
	List(opts metav1.ListOptions) (runtime.Object, error)
	// Get gets the resource with the specified name.
	Get(name string, opts metav1.GetOptions) (*unstructured.Unstructured, error)
	// Delete deletes the resource with the specified name.
	Delete(name string, opts *metav1.DeleteOptions) error
	// DeleteCollection deletes a collection of objects.
	DeleteCollection(deleteOptions *metav1.DeleteOptions, listOptions metav1.ListOptions) error
	// Create creates the provided resource.
	Create(obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	// Update updates the provided resource.
	Update(obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	// Watch returns a watch.Interface that watches the resource.
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	// Patch patches the provided resource.
	Patch(name string, pt types.PatchType, data []byte) (*unstructured.Unstructured, error)
}

// Client is a Kubernetes client that allows you to access metadata
// and manipulate metadata of a Kubernetes API group, and implements Interface.
type Client struct {
	version  schema.GroupVersion
	delegate DynamicInterface
}

// NewClient returns a new client based on the passed in config. The
// codec is ignored, as the dynamic client uses it's own codec.
func NewClient(conf *restclient.Config, version schema.GroupVersion) (*Client, error) {
	delegate, err := NewForConfig(conf)
	if err != nil {
		return nil, err
	}

	return &Client{version: version, delegate: delegate}, nil
}

// Resource returns an API interface to the specified resource for this client's
// group and version. If resource is not a namespaced resource, then namespace
// is ignored. The ResourceInterface inherits the parameter codec of c.
func (c *Client) Resource(resource *metav1.APIResource, namespace string) ResourceInterface {
	resourceTokens := strings.SplitN(resource.Name, "/", 2)
	subresource := ""
	if len(resourceTokens) > 1 {
		subresource = resourceTokens[1]
	}

	if len(namespace) == 0 {
		return oldResourceShim(c.delegate.ClusterSubresource(c.version.WithResource(resourceTokens[0]), subresource))
	}
	return oldResourceShim(c.delegate.NamespacedSubresource(c.version.WithResource(resourceTokens[0]), subresource, namespace))
}

// the old interfaces used the wrong type for lists.  this fixes that
func oldResourceShim(in DynamicResourceInterface) ResourceInterface {
	return oldResourceShimType{DynamicResourceInterface: in}
}

type oldResourceShimType struct {
	DynamicResourceInterface
}

func (s oldResourceShimType) List(opts metav1.ListOptions) (runtime.Object, error) {
	return s.DynamicResourceInterface.List(opts)
}

func (s oldResourceShimType) Patch(name string, pt types.PatchType, data []byte) (*unstructured.Unstructured, error) {
	return s.DynamicResourceInterface.Patch(name, pt, data)
}

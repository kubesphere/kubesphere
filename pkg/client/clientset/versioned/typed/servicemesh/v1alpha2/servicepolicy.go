/*
Copyright 2020 The KubeSphere Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"

	v1alpha2 "kubesphere.io/kubesphere/pkg/apis/servicemesh/v1alpha2"
	scheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
)

// ServicePoliciesGetter has a method to return a ServicePolicyInterface.
// A group's client should implement this interface.
type ServicePoliciesGetter interface {
	ServicePolicies(namespace string) ServicePolicyInterface
}

// ServicePolicyInterface has methods to work with ServicePolicy resources.
type ServicePolicyInterface interface {
	Create(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.CreateOptions) (*v1alpha2.ServicePolicy, error)
	Update(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.UpdateOptions) (*v1alpha2.ServicePolicy, error)
	UpdateStatus(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.UpdateOptions) (*v1alpha2.ServicePolicy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.ServicePolicy, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.ServicePolicyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.ServicePolicy, err error)
	ServicePolicyExpansion
}

// servicePolicies implements ServicePolicyInterface
type servicePolicies struct {
	client rest.Interface
	ns     string
}

// newServicePolicies returns a ServicePolicies
func newServicePolicies(c *ServicemeshV1alpha2Client, namespace string) *servicePolicies {
	return &servicePolicies{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the servicePolicy, and returns the corresponding servicePolicy object, and an error if there is any.
func (c *servicePolicies) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.ServicePolicy, err error) {
	result = &v1alpha2.ServicePolicy{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicepolicies").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServicePolicies that match those selectors.
func (c *servicePolicies) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.ServicePolicyList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.ServicePolicyList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("servicepolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested servicePolicies.
func (c *servicePolicies) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("servicepolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a servicePolicy and creates it.  Returns the server's representation of the servicePolicy, and an error, if there is any.
func (c *servicePolicies) Create(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.CreateOptions) (result *v1alpha2.ServicePolicy, err error) {
	result = &v1alpha2.ServicePolicy{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("servicepolicies").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(servicePolicy).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a servicePolicy and updates it. Returns the server's representation of the servicePolicy, and an error, if there is any.
func (c *servicePolicies) Update(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.UpdateOptions) (result *v1alpha2.ServicePolicy, err error) {
	result = &v1alpha2.ServicePolicy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicepolicies").
		Name(servicePolicy.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(servicePolicy).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *servicePolicies) UpdateStatus(ctx context.Context, servicePolicy *v1alpha2.ServicePolicy, opts v1.UpdateOptions) (result *v1alpha2.ServicePolicy, err error) {
	result = &v1alpha2.ServicePolicy{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("servicepolicies").
		Name(servicePolicy.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(servicePolicy).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the servicePolicy and deletes it. Returns an error if one occurs.
func (c *servicePolicies) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicepolicies").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *servicePolicies) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("servicepolicies").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched servicePolicy.
func (c *servicePolicies) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.ServicePolicy, err error) {
	result = &v1alpha2.ServicePolicy{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("servicepolicies").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}

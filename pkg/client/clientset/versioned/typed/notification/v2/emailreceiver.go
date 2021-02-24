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

package v2

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v2 "kubesphere.io/kubesphere/pkg/apis/notification/v2"
	scheme "kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
)

// EmailReceiversGetter has a method to return a EmailReceiverInterface.
// A group's client should implement this interface.
type EmailReceiversGetter interface {
	EmailReceivers() EmailReceiverInterface
}

// EmailReceiverInterface has methods to work with EmailReceiver resources.
type EmailReceiverInterface interface {
	Create(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.CreateOptions) (*v2.EmailReceiver, error)
	Update(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.UpdateOptions) (*v2.EmailReceiver, error)
	UpdateStatus(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.UpdateOptions) (*v2.EmailReceiver, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v2.EmailReceiver, error)
	List(ctx context.Context, opts v1.ListOptions) (*v2.EmailReceiverList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2.EmailReceiver, err error)
	EmailReceiverExpansion
}

// emailReceivers implements EmailReceiverInterface
type emailReceivers struct {
	client rest.Interface
}

// newEmailReceivers returns a EmailReceivers
func newEmailReceivers(c *NotificationV2Client) *emailReceivers {
	return &emailReceivers{
		client: c.RESTClient(),
	}
}

// Get takes name of the emailReceiver, and returns the corresponding emailReceiver object, and an error if there is any.
func (c *emailReceivers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v2.EmailReceiver, err error) {
	result = &v2.EmailReceiver{}
	err = c.client.Get().
		Resource("emailreceivers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of EmailReceivers that match those selectors.
func (c *emailReceivers) List(ctx context.Context, opts v1.ListOptions) (result *v2.EmailReceiverList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v2.EmailReceiverList{}
	err = c.client.Get().
		Resource("emailreceivers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested emailReceivers.
func (c *emailReceivers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("emailreceivers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a emailReceiver and creates it.  Returns the server's representation of the emailReceiver, and an error, if there is any.
func (c *emailReceivers) Create(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.CreateOptions) (result *v2.EmailReceiver, err error) {
	result = &v2.EmailReceiver{}
	err = c.client.Post().
		Resource("emailreceivers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(emailReceiver).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a emailReceiver and updates it. Returns the server's representation of the emailReceiver, and an error, if there is any.
func (c *emailReceivers) Update(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.UpdateOptions) (result *v2.EmailReceiver, err error) {
	result = &v2.EmailReceiver{}
	err = c.client.Put().
		Resource("emailreceivers").
		Name(emailReceiver.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(emailReceiver).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *emailReceivers) UpdateStatus(ctx context.Context, emailReceiver *v2.EmailReceiver, opts v1.UpdateOptions) (result *v2.EmailReceiver, err error) {
	result = &v2.EmailReceiver{}
	err = c.client.Put().
		Resource("emailreceivers").
		Name(emailReceiver.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(emailReceiver).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the emailReceiver and deletes it. Returns an error if one occurs.
func (c *emailReceivers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("emailreceivers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *emailReceivers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("emailreceivers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched emailReceiver.
func (c *emailReceivers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2.EmailReceiver, err error) {
	result = &v2.EmailReceiver{}
	err = c.client.Patch(pt).
		Resource("emailreceivers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
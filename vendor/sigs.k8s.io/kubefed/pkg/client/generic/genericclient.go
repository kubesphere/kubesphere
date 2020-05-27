/*
Copyright 2019 The Kubernetes Authors.

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

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"sigs.k8s.io/kubefed/pkg/client/generic/scheme"
)

type Client interface {
	Create(ctx context.Context, obj runtime.Object) error
	Get(ctx context.Context, obj runtime.Object, namespace, name string) error
	Update(ctx context.Context, obj runtime.Object) error
	Delete(ctx context.Context, obj runtime.Object, namespace, name string) error
	List(ctx context.Context, obj runtime.Object, namespace string, opts ...client.ListOption) error
	UpdateStatus(ctx context.Context, obj runtime.Object) error
}

type genericClient struct {
	client client.Client
}

func New(config *rest.Config) (Client, error) {
	client, err := client.New(config, client.Options{Scheme: scheme.Scheme})
	return &genericClient{client}, err
}

func NewForConfigOrDie(config *rest.Config) Client {
	client, err := New(config)
	if err != nil {
		panic(err)
	}
	return client
}

func NewForConfigOrDieWithUserAgent(config *rest.Config, userAgent string) Client {
	configCopy := rest.CopyConfig(config)
	rest.AddUserAgent(configCopy, userAgent)
	return NewForConfigOrDie(configCopy)
}

func (c *genericClient) Create(ctx context.Context, obj runtime.Object) error {
	return c.client.Create(ctx, obj)
}

func (c *genericClient) Get(ctx context.Context, obj runtime.Object, namespace, name string) error {
	return c.client.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, obj)
}

func (c *genericClient) Update(ctx context.Context, obj runtime.Object) error {
	return c.client.Update(ctx, obj)
}

func (c *genericClient) Delete(ctx context.Context, obj runtime.Object, namespace, name string) error {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	accessor.SetNamespace(namespace)
	accessor.SetName(name)
	return c.client.Delete(ctx, obj)
}

func (c *genericClient) List(ctx context.Context, obj runtime.Object, namespace string, opts ...client.ListOption) error {
	opts = append(opts, client.InNamespace(namespace))
	return c.client.List(ctx, obj, opts...)
}

func (c *genericClient) UpdateStatus(ctx context.Context, obj runtime.Object) error {
	return c.client.Status().Update(ctx, obj)
}

/*
Copyright 2014 The Kubernetes Authors.

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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
)

type ClientConfigFunc func() (*rest.Config, error)

// RESTClient is a client helper for dealing with RESTful resources
// in a generic way.
type RESTClient interface {
	Get() *rest.Request
	Post() *rest.Request
	Patch(types.PatchType) *rest.Request
	Delete() *rest.Request
	Put() *rest.Request
}

// RequestTransform is a function that is given a chance to modify the outgoing request.
type RequestTransform func(*rest.Request)

// NewClientWithOptions wraps the provided RESTClient and invokes each transform on each
// newly created request.
func NewClientWithOptions(c RESTClient, transforms ...RequestTransform) RESTClient {
	return &clientOptions{c: c, transforms: transforms}
}

type clientOptions struct {
	c          RESTClient
	transforms []RequestTransform
}

func (c *clientOptions) modify(req *rest.Request) *rest.Request {
	for _, transform := range c.transforms {
		transform(req)
	}
	return req
}

func (c *clientOptions) Get() *rest.Request {
	return c.modify(c.c.Get())
}

func (c *clientOptions) Post() *rest.Request {
	return c.modify(c.c.Post())
}
func (c *clientOptions) Patch(t types.PatchType) *rest.Request {
	return c.modify(c.c.Patch(t))
}
func (c *clientOptions) Delete() *rest.Request {
	return c.modify(c.c.Delete())
}
func (c *clientOptions) Put() *rest.Request {
	return c.modify(c.c.Put())
}

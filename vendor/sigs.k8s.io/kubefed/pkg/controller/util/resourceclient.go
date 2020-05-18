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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type ResourceClient interface {
	Resources(namespace string) dynamic.ResourceInterface
	Kind() string
}

type resourceClient struct {
	client      dynamic.Interface
	apiResource schema.GroupVersionResource
	namespaced  bool
	kind        string
}

func NewResourceClient(config *rest.Config, apiResource *metav1.APIResource) (ResourceClient, error) {
	resource := schema.GroupVersionResource{
		Group:    apiResource.Group,
		Version:  apiResource.Version,
		Resource: apiResource.Name,
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &resourceClient{
		client:      client,
		apiResource: resource,
		namespaced:  apiResource.Namespaced,
		kind:        apiResource.Kind,
	}, nil
}

func (c *resourceClient) Resources(namespace string) dynamic.ResourceInterface {
	// TODO(marun) Consider returning Interface instead of
	// ResourceInterface to allow callers to decide if they want to
	// invoke Namespace().  Either that, or replace the use of
	// ResourceClient with the controller-runtime generic client.
	if c.namespaced {
		return c.client.Resource(c.apiResource).Namespace(namespace)
	}
	return c.client.Resource(c.apiResource)
}

func (c *resourceClient) Kind() string {
	return c.kind
}

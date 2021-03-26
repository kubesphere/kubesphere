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

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"

	v2beta1 "kubesphere.io/kubesphere/pkg/apis/notification/v2beta1"
)

// FakeConfigs implements ConfigInterface
type FakeConfigs struct {
	Fake *FakeNotificationV2beta1
}

var configsResource = schema.GroupVersionResource{Group: "notification.kubesphere.io", Version: "v2beta1", Resource: "configs"}

var configsKind = schema.GroupVersionKind{Group: "notification.kubesphere.io", Version: "v2beta1", Kind: "Config"}

// Get takes name of the config, and returns the corresponding config object, and an error if there is any.
func (c *FakeConfigs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v2beta1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(configsResource, name), &v2beta1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta1.Config), err
}

// List takes label and field selectors, and returns the list of Configs that match those selectors.
func (c *FakeConfigs) List(ctx context.Context, opts v1.ListOptions) (result *v2beta1.ConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(configsResource, configsKind, opts), &v2beta1.ConfigList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v2beta1.ConfigList{ListMeta: obj.(*v2beta1.ConfigList).ListMeta}
	for _, item := range obj.(*v2beta1.ConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested configs.
func (c *FakeConfigs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(configsResource, opts))
}

// Create takes the representation of a config and creates it.  Returns the server's representation of the config, and an error, if there is any.
func (c *FakeConfigs) Create(ctx context.Context, config *v2beta1.Config, opts v1.CreateOptions) (result *v2beta1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(configsResource, config), &v2beta1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta1.Config), err
}

// Update takes the representation of a config and updates it. Returns the server's representation of the config, and an error, if there is any.
func (c *FakeConfigs) Update(ctx context.Context, config *v2beta1.Config, opts v1.UpdateOptions) (result *v2beta1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(configsResource, config), &v2beta1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta1.Config), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeConfigs) UpdateStatus(ctx context.Context, config *v2beta1.Config, opts v1.UpdateOptions) (*v2beta1.Config, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(configsResource, "status", config), &v2beta1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta1.Config), err
}

// Delete takes name of the config and deletes it. Returns an error if one occurs.
func (c *FakeConfigs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(configsResource, name), &v2beta1.Config{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeConfigs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(configsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v2beta1.ConfigList{})
	return err
}

// Patch applies the patch and returns the patched config.
func (c *FakeConfigs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v2beta1.Config, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(configsResource, name, pt, data, subresources...), &v2beta1.Config{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v2beta1.Config), err
}

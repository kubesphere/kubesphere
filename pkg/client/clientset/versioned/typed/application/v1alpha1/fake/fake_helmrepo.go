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
	v1alpha1 "kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
)

// FakeHelmRepos implements HelmRepoInterface
type FakeHelmRepos struct {
	Fake *FakeApplicationV1alpha1
}

var helmreposResource = schema.GroupVersionResource{Group: "application.kubesphere.io", Version: "v1alpha1", Resource: "helmrepos"}

var helmreposKind = schema.GroupVersionKind{Group: "application.kubesphere.io", Version: "v1alpha1", Kind: "HelmRepo"}

// Get takes name of the helmRepo, and returns the corresponding helmRepo object, and an error if there is any.
func (c *FakeHelmRepos) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.HelmRepo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(helmreposResource, name), &v1alpha1.HelmRepo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmRepo), err
}

// List takes label and field selectors, and returns the list of HelmRepos that match those selectors.
func (c *FakeHelmRepos) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.HelmRepoList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(helmreposResource, helmreposKind, opts), &v1alpha1.HelmRepoList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.HelmRepoList{ListMeta: obj.(*v1alpha1.HelmRepoList).ListMeta}
	for _, item := range obj.(*v1alpha1.HelmRepoList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested helmRepos.
func (c *FakeHelmRepos) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(helmreposResource, opts))
}

// Create takes the representation of a helmRepo and creates it.  Returns the server's representation of the helmRepo, and an error, if there is any.
func (c *FakeHelmRepos) Create(ctx context.Context, helmRepo *v1alpha1.HelmRepo, opts v1.CreateOptions) (result *v1alpha1.HelmRepo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(helmreposResource, helmRepo), &v1alpha1.HelmRepo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmRepo), err
}

// Update takes the representation of a helmRepo and updates it. Returns the server's representation of the helmRepo, and an error, if there is any.
func (c *FakeHelmRepos) Update(ctx context.Context, helmRepo *v1alpha1.HelmRepo, opts v1.UpdateOptions) (result *v1alpha1.HelmRepo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(helmreposResource, helmRepo), &v1alpha1.HelmRepo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmRepo), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeHelmRepos) UpdateStatus(ctx context.Context, helmRepo *v1alpha1.HelmRepo, opts v1.UpdateOptions) (*v1alpha1.HelmRepo, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(helmreposResource, "status", helmRepo), &v1alpha1.HelmRepo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmRepo), err
}

// Delete takes name of the helmRepo and deletes it. Returns an error if one occurs.
func (c *FakeHelmRepos) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(helmreposResource, name), &v1alpha1.HelmRepo{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeHelmRepos) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(helmreposResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.HelmRepoList{})
	return err
}

// Patch applies the patch and returns the patched helmRepo.
func (c *FakeHelmRepos) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.HelmRepo, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(helmreposResource, name, pt, data, subresources...), &v1alpha1.HelmRepo{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.HelmRepo), err
}

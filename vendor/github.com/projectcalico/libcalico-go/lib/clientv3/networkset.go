// Copyright (c) 2018-2019 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clientv3

import (
	"context"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/options"
	validator "github.com/projectcalico/libcalico-go/lib/validator/v3"
	"github.com/projectcalico/libcalico-go/lib/watch"
)

// NetworkSetInterface has methods to work with NetworkSet resources.
type NetworkSetInterface interface {
	Create(ctx context.Context, res *apiv3.NetworkSet, opts options.SetOptions) (*apiv3.NetworkSet, error)
	Update(ctx context.Context, res *apiv3.NetworkSet, opts options.SetOptions) (*apiv3.NetworkSet, error)
	Delete(ctx context.Context, namespace, name string, opts options.DeleteOptions) (*apiv3.NetworkSet, error)
	Get(ctx context.Context, namespace, name string, opts options.GetOptions) (*apiv3.NetworkSet, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.NetworkSetList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// networkSets implements NetworkSetInterface
type networkSets struct {
	client client
}

// Create takes the representation of a NetworkSet and creates it.  Returns the stored
// representation of the NetworkSet, and an error, if there is any.
func (r networkSets) Create(ctx context.Context, res *apiv3.NetworkSet, opts options.SetOptions) (*apiv3.NetworkSet, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}
	out, err := r.client.resources.Create(ctx, opts, apiv3.KindNetworkSet, res)
	if out != nil {
		return out.(*apiv3.NetworkSet), err
	}
	return nil, err
}

// Update takes the representation of a NetworkSet and updates it. Returns the stored
// representation of the NetworkSet, and an error, if there is any.
func (r networkSets) Update(ctx context.Context, res *apiv3.NetworkSet, opts options.SetOptions) (*apiv3.NetworkSet, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}
	out, err := r.client.resources.Update(ctx, opts, apiv3.KindNetworkSet, res)
	if out != nil {
		return out.(*apiv3.NetworkSet), err
	}
	return nil, err
}

// Delete takes name of the NetworkSet and deletes it. Returns an error if one occurs.
func (r networkSets) Delete(ctx context.Context, namespace, name string, opts options.DeleteOptions) (*apiv3.NetworkSet, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindNetworkSet, namespace, name)
	if out != nil {
		return out.(*apiv3.NetworkSet), err
	}
	return nil, err
}

// Get takes name of the NetworkSet, and returns the corresponding NetworkSet object,
// and an error if there is any.
func (r networkSets) Get(ctx context.Context, namespace, name string, opts options.GetOptions) (*apiv3.NetworkSet, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindNetworkSet, namespace, name)
	if out != nil {
		return out.(*apiv3.NetworkSet), err
	}
	return nil, err
}

// List returns the list of NetworkSet objects that match the supplied options.
func (r networkSets) List(ctx context.Context, opts options.ListOptions) (*apiv3.NetworkSetList, error) {
	res := &apiv3.NetworkSetList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindNetworkSet, apiv3.KindNetworkSetList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the NetworkSets that match the
// supplied options.
func (r networkSets) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindNetworkSet, nil)
}

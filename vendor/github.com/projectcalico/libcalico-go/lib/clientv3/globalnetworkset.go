// Copyright (c) 2018 Tigera, Inc. All rights reserved.

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

// GlobalNetworkSetInterface has methods to work with GlobalNetworkSet resources.
type GlobalNetworkSetInterface interface {
	Create(ctx context.Context, res *apiv3.GlobalNetworkSet, opts options.SetOptions) (*apiv3.GlobalNetworkSet, error)
	Update(ctx context.Context, res *apiv3.GlobalNetworkSet, opts options.SetOptions) (*apiv3.GlobalNetworkSet, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.GlobalNetworkSet, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.GlobalNetworkSet, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.GlobalNetworkSetList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// globalNetworkSets implements GlobalNetworkSetInterface
type globalNetworkSets struct {
	client client
}

// Create takes the representation of a GlobalNetworkSet and creates it.  Returns the stored
// representation of the GlobalNetworkSet, and an error, if there is any.
func (r globalNetworkSets) Create(ctx context.Context, res *apiv3.GlobalNetworkSet, opts options.SetOptions) (*apiv3.GlobalNetworkSet, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindGlobalNetworkSet, res)
	if out != nil {
		return out.(*apiv3.GlobalNetworkSet), err
	}
	return nil, err
}

// Update takes the representation of a GlobalNetworkSet and updates it. Returns the stored
// representation of the GlobalNetworkSet, and an error, if there is any.
func (r globalNetworkSets) Update(ctx context.Context, res *apiv3.GlobalNetworkSet, opts options.SetOptions) (*apiv3.GlobalNetworkSet, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindGlobalNetworkSet, res)
	if out != nil {
		return out.(*apiv3.GlobalNetworkSet), err
	}
	return nil, err
}

// Delete takes name of the GlobalNetworkSet and deletes it. Returns an error if one occurs.
func (r globalNetworkSets) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.GlobalNetworkSet, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindGlobalNetworkSet, noNamespace, name)
	if out != nil {
		return out.(*apiv3.GlobalNetworkSet), err
	}
	return nil, err
}

// Get takes name of the GlobalNetworkSet, and returns the corresponding GlobalNetworkSet object,
// and an error if there is any.
func (r globalNetworkSets) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.GlobalNetworkSet, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindGlobalNetworkSet, noNamespace, name)
	if out != nil {
		return out.(*apiv3.GlobalNetworkSet), err
	}
	return nil, err
}

// List returns the list of GlobalNetworkSet objects that match the supplied options.
func (r globalNetworkSets) List(ctx context.Context, opts options.ListOptions) (*apiv3.GlobalNetworkSetList, error) {
	res := &apiv3.GlobalNetworkSetList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindGlobalNetworkSet, apiv3.KindGlobalNetworkSetList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the GlobalNetworkSets that match the
// supplied options.
func (r globalNetworkSets) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindGlobalNetworkSet, nil)
}

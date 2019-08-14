// Copyright (c) 2017 Tigera, Inc. All rights reserved.

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

// HostEndpointInterface has methods to work with HostEndpoint resources.
type HostEndpointInterface interface {
	Create(ctx context.Context, res *apiv3.HostEndpoint, opts options.SetOptions) (*apiv3.HostEndpoint, error)
	Update(ctx context.Context, res *apiv3.HostEndpoint, opts options.SetOptions) (*apiv3.HostEndpoint, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.HostEndpoint, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.HostEndpoint, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.HostEndpointList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// hostEndpoints implements HostEndpointInterface
type hostEndpoints struct {
	client client
}

// Create takes the representation of a HostEndpoint and creates it.  Returns the stored
// representation of the HostEndpoint, and an error, if there is any.
func (r hostEndpoints) Create(ctx context.Context, res *apiv3.HostEndpoint, opts options.SetOptions) (*apiv3.HostEndpoint, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindHostEndpoint, res)
	if out != nil {
		return out.(*apiv3.HostEndpoint), err
	}
	return nil, err
}

// Update takes the representation of a HostEndpoint and updates it. Returns the stored
// representation of the HostEndpoint, and an error, if there is any.
func (r hostEndpoints) Update(ctx context.Context, res *apiv3.HostEndpoint, opts options.SetOptions) (*apiv3.HostEndpoint, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindHostEndpoint, res)
	if out != nil {
		return out.(*apiv3.HostEndpoint), err
	}
	return nil, err
}

// Delete takes name of the HostEndpoint and deletes it. Returns an error if one occurs.
func (r hostEndpoints) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.HostEndpoint, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindHostEndpoint, noNamespace, name)
	if out != nil {
		return out.(*apiv3.HostEndpoint), err
	}
	return nil, err
}

// Get takes name of the HostEndpoint, and returns the corresponding HostEndpoint object,
// and an error if there is any.
func (r hostEndpoints) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.HostEndpoint, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindHostEndpoint, noNamespace, name)
	if out != nil {
		return out.(*apiv3.HostEndpoint), err
	}
	return nil, err
}

// List returns the list of HostEndpoint objects that match the supplied options.
func (r hostEndpoints) List(ctx context.Context, opts options.ListOptions) (*apiv3.HostEndpointList, error) {
	res := &apiv3.HostEndpointList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindHostEndpoint, apiv3.KindHostEndpointList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the HostEndpoints that match the
// supplied options.
func (r hostEndpoints) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindHostEndpoint, nil)
}

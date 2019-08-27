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

// ProfileInterface has methods to work with Profile resources.
type ProfileInterface interface {
	Create(ctx context.Context, res *apiv3.Profile, opts options.SetOptions) (*apiv3.Profile, error)
	Update(ctx context.Context, res *apiv3.Profile, opts options.SetOptions) (*apiv3.Profile, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.Profile, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.Profile, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.ProfileList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// profiles implements ProfileInterface
type profiles struct {
	client client
}

// Create takes the representation of a Profile and creates it.  Returns the stored
// representation of the Profile, and an error, if there is any.
func (r profiles) Create(ctx context.Context, res *apiv3.Profile, opts options.SetOptions) (*apiv3.Profile, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindProfile, res)
	if out != nil {
		return out.(*apiv3.Profile), err
	}
	return nil, err
}

// Update takes the representation of a Profile and updates it. Returns the stored
// representation of the Profile, and an error, if there is any.
func (r profiles) Update(ctx context.Context, res *apiv3.Profile, opts options.SetOptions) (*apiv3.Profile, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindProfile, res)
	if out != nil {
		return out.(*apiv3.Profile), err
	}
	return nil, err
}

// Delete takes name of the Profile and deletes it. Returns an error if one occurs.
func (r profiles) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.Profile, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindProfile, noNamespace, name)
	if out != nil {
		return out.(*apiv3.Profile), err
	}
	return nil, err
}

// Get takes name of the Profile, and returns the corresponding Profile object,
// and an error if there is any.
func (r profiles) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.Profile, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindProfile, noNamespace, name)
	if out != nil {
		return out.(*apiv3.Profile), err
	}
	return nil, err
}

// List returns the list of Profile objects that match the supplied options.
func (r profiles) List(ctx context.Context, opts options.ListOptions) (*apiv3.ProfileList, error) {
	res := &apiv3.ProfileList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindProfile, apiv3.KindProfileList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the Profiles that match the
// supplied options.
func (r profiles) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindProfile, nil)
}

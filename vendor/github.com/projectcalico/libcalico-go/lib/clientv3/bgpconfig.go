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
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/options"
	validator "github.com/projectcalico/libcalico-go/lib/validator/v3"
	"github.com/projectcalico/libcalico-go/lib/watch"
)

// BGPConfigurationInterface has methods to work with BGPConfiguration resources.
type BGPConfigurationInterface interface {
	Create(ctx context.Context, res *apiv3.BGPConfiguration, opts options.SetOptions) (*apiv3.BGPConfiguration, error)
	Update(ctx context.Context, res *apiv3.BGPConfiguration, opts options.SetOptions) (*apiv3.BGPConfiguration, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.BGPConfiguration, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.BGPConfiguration, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.BGPConfigurationList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// bgpConfigurations implements BGPConfigurationInterface
type bgpConfigurations struct {
	client client
}

// Create takes the representation of a BGPConfiguration and creates it.
// Returns the stored representation of the BGPConfiguration, and an error
// if there is any.
func (r bgpConfigurations) Create(ctx context.Context, res *apiv3.BGPConfiguration, opts options.SetOptions) (*apiv3.BGPConfiguration, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	if err := r.ValidateDefaultOnlyFields(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindBGPConfiguration, res)
	if out != nil {
		return out.(*apiv3.BGPConfiguration), err
	}
	return nil, err
}

// Update takes the representation of a BGPConfiguration and updates it.
// Returns the stored representation of the BGPConfiguration, and an error
// if there is any.
func (r bgpConfigurations) Update(ctx context.Context, res *apiv3.BGPConfiguration, opts options.SetOptions) (*apiv3.BGPConfiguration, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// Check that NodeToNodeMeshEnabled and ASNumber are set. Can only be set on "default".
	if err := r.ValidateDefaultOnlyFields(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindBGPConfiguration, res)
	if out != nil {
		return out.(*apiv3.BGPConfiguration), err
	}
	return nil, err
}

// Delete takes name of the BGPConfiguration and deletes it. Returns an
// error if one occurs.
func (r bgpConfigurations) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.BGPConfiguration, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindBGPConfiguration, noNamespace, name)
	if out != nil {
		return out.(*apiv3.BGPConfiguration), err
	}
	return nil, err
}

// Get takes name of the BGPConfiguration, and returns the corresponding
// BGPConfiguration object, and an error if there is any.
func (r bgpConfigurations) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.BGPConfiguration, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindBGPConfiguration, noNamespace, name)
	if out != nil {
		return out.(*apiv3.BGPConfiguration), err
	}
	return nil, err
}

// List returns the list of BGPConfiguration objects that match the supplied options.
func (r bgpConfigurations) List(ctx context.Context, opts options.ListOptions) (*apiv3.BGPConfigurationList, error) {
	res := &apiv3.BGPConfigurationList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindBGPConfiguration, apiv3.KindBGPConfigurationList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the BGPConfiguration that
// match the supplied options.
func (r bgpConfigurations) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindBGPConfiguration, nil)
}

func (r bgpConfigurations) ValidateDefaultOnlyFields(res *apiv3.BGPConfiguration) error {
	errFields := []cerrors.ErroredField{}
	if res.ObjectMeta.GetName() != "default" {
		if res.Spec.NodeToNodeMeshEnabled != nil {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "BGPConfiguration.Spec.NodeToNodeMeshEnabled",
				Reason: "Cannot set nodeToNodeMeshEnabled on a non default BGP Configuration.",
			})
		}

		if res.Spec.ASNumber != nil {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "BGPConfiguration.Spec.ASNumber",
				Reason: "Cannot set ASNumber on a non default BGP Configuration.",
			})
		}

		if res.Spec.ServiceExternalIPs != nil && len(res.Spec.ServiceExternalIPs) > 0 {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "BGPConfiguration.Spec.ServiceExternalIPs",
				Reason: "Cannot set ServiceExternalIPs on a non default BGP Configuration.",
			})
		}

		if res.Spec.ServiceClusterIPs != nil && len(res.Spec.ServiceClusterIPs) > 0 {
			errFields = append(errFields, cerrors.ErroredField{
				Name:   "BGPConfiguration.Spec.ServiceClusterIPs",
				Reason: "Cannot set ServiceClusterIPs on a non default BGP Configuration.",
			})
		}
	}

	if len(errFields) > 0 {
		return cerrors.ErrorValidation{
			ErroredFields: errFields,
		}
	}

	return nil
}

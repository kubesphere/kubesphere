// Copyright (c) 2017-2018 Tigera, Inc. All rights reserved.

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

// NetworkPolicyInterface has methods to work with NetworkPolicy resources.
type NetworkPolicyInterface interface {
	Create(ctx context.Context, res *apiv3.NetworkPolicy, opts options.SetOptions) (*apiv3.NetworkPolicy, error)
	Update(ctx context.Context, res *apiv3.NetworkPolicy, opts options.SetOptions) (*apiv3.NetworkPolicy, error)
	Delete(ctx context.Context, namespace, name string, opts options.DeleteOptions) (*apiv3.NetworkPolicy, error)
	Get(ctx context.Context, namespace, name string, opts options.GetOptions) (*apiv3.NetworkPolicy, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.NetworkPolicyList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// networkPolicies implements NetworkPolicyInterface
type networkPolicies struct {
	client client
}

// Create takes the representation of a NetworkPolicy and creates it.  Returns the stored
// representation of the NetworkPolicy, and an error, if there is any.
func (r networkPolicies) Create(ctx context.Context, res *apiv3.NetworkPolicy, opts options.SetOptions) (*apiv3.NetworkPolicy, error) {
	if res != nil {
		// Since we're about to default some fields, take a (shallow) copy of the input data
		// before we do so.
		resCopy := *res
		res = &resCopy
	}
	defaultPolicyTypesField(res.Spec.Ingress, res.Spec.Egress, &res.Spec.Types)

	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// Properly prefix the name
	res.GetObjectMeta().SetName(convertPolicyNameForStorage(res.GetObjectMeta().GetName()))
	out, err := r.client.resources.Create(ctx, opts, apiv3.KindNetworkPolicy, res)
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.NetworkPolicy), err
	}

	// Remove the prefix out of the returned policy name.
	res.GetObjectMeta().SetName(convertPolicyNameFromStorage(res.GetObjectMeta().GetName()))
	return nil, err
}

// Update takes the representation of a NetworkPolicy and updates it. Returns the stored
// representation of the NetworkPolicy, and an error, if there is any.
func (r networkPolicies) Update(ctx context.Context, res *apiv3.NetworkPolicy, opts options.SetOptions) (*apiv3.NetworkPolicy, error) {
	if res != nil {
		// Since we're about to default some fields, take a (shallow) copy of the input data
		// before we do so.
		resCopy := *res
		res = &resCopy
	}
	defaultPolicyTypesField(res.Spec.Ingress, res.Spec.Egress, &res.Spec.Types)

	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// Properly prefix the name
	res.GetObjectMeta().SetName(convertPolicyNameForStorage(res.GetObjectMeta().GetName()))
	out, err := r.client.resources.Update(ctx, opts, apiv3.KindNetworkPolicy, res)
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.NetworkPolicy), err
	}

	// Remove the prefix out of the returned policy name.
	res.GetObjectMeta().SetName(convertPolicyNameFromStorage(res.GetObjectMeta().GetName()))
	return nil, err
}

// Delete takes name of the NetworkPolicy and deletes it. Returns an error if one occurs.
func (r networkPolicies) Delete(ctx context.Context, namespace, name string, opts options.DeleteOptions) (*apiv3.NetworkPolicy, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindNetworkPolicy, namespace, convertPolicyNameForStorage(name))
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.NetworkPolicy), err
	}
	return nil, err
}

// Get takes name of the NetworkPolicy, and returns the corresponding NetworkPolicy object,
// and an error if there is any.
func (r networkPolicies) Get(ctx context.Context, namespace, name string, opts options.GetOptions) (*apiv3.NetworkPolicy, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindNetworkPolicy, namespace, convertPolicyNameForStorage(name))
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.NetworkPolicy), err
	}
	return nil, err
}

// List returns the list of NetworkPolicy objects that match the supplied options.
func (r networkPolicies) List(ctx context.Context, opts options.ListOptions) (*apiv3.NetworkPolicyList, error) {
	res := &apiv3.NetworkPolicyList{}
	// Add the name prefix if name is provided
	if opts.Name != "" {
		opts.Name = convertPolicyNameForStorage(opts.Name)
	}

	if err := r.client.resources.List(ctx, opts, apiv3.KindNetworkPolicy, apiv3.KindNetworkPolicyList, res); err != nil {
		return nil, err
	}

	// Remove the prefix off of each policy name
	for i, _ := range res.Items {
		name := res.Items[i].GetObjectMeta().GetName()
		res.Items[i].GetObjectMeta().SetName(convertPolicyNameFromStorage(name))
	}

	return res, nil
}

// Watch returns a watch.Interface that watches the NetworkPolicies that match the
// supplied options.
func (r networkPolicies) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	// Add the name prefix if name is provided
	if opts.Name != "" {
		opts.Name = convertPolicyNameForStorage(opts.Name)
	}

	return r.client.resources.Watch(ctx, opts, apiv3.KindNetworkPolicy, &policyConverter{})
}

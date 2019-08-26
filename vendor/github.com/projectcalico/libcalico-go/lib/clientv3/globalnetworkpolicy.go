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
	"strings"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/options"
	validator "github.com/projectcalico/libcalico-go/lib/validator/v3"
	"github.com/projectcalico/libcalico-go/lib/watch"
)

// GlobalNetworkPolicyInterface has methods to work with GlobalNetworkPolicy resources.
type GlobalNetworkPolicyInterface interface {
	Create(ctx context.Context, res *apiv3.GlobalNetworkPolicy, opts options.SetOptions) (*apiv3.GlobalNetworkPolicy, error)
	Update(ctx context.Context, res *apiv3.GlobalNetworkPolicy, opts options.SetOptions) (*apiv3.GlobalNetworkPolicy, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.GlobalNetworkPolicy, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.GlobalNetworkPolicy, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.GlobalNetworkPolicyList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// globalNetworkPolicies implements GlobalNetworkPolicyInterface
type globalNetworkPolicies struct {
	client client
}

// Create takes the representation of a GlobalNetworkPolicy and creates it.  Returns the stored
// representation of the GlobalNetworkPolicy, and an error, if there is any.
func (r globalNetworkPolicies) Create(ctx context.Context, res *apiv3.GlobalNetworkPolicy, opts options.SetOptions) (*apiv3.GlobalNetworkPolicy, error) {
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
	out, err := r.client.resources.Create(ctx, opts, apiv3.KindGlobalNetworkPolicy, res)
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.GlobalNetworkPolicy), err
	}

	// Remove the prefix out of the returned policy name.
	res.GetObjectMeta().SetName(convertPolicyNameFromStorage(res.GetObjectMeta().GetName()))
	return nil, err
}

// Update takes the representation of a GlobalNetworkPolicy and updates it. Returns the stored
// representation of the GlobalNetworkPolicy, and an error, if there is any.
func (r globalNetworkPolicies) Update(ctx context.Context, res *apiv3.GlobalNetworkPolicy, opts options.SetOptions) (*apiv3.GlobalNetworkPolicy, error) {
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
	out, err := r.client.resources.Update(ctx, opts, apiv3.KindGlobalNetworkPolicy, res)
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.GlobalNetworkPolicy), err
	}

	// Remove the prefix out of the returned policy name.
	res.GetObjectMeta().SetName(convertPolicyNameFromStorage(res.GetObjectMeta().GetName()))
	return nil, err
}

// Delete takes name of the GlobalNetworkPolicy and deletes it. Returns an error if one occurs.
func (r globalNetworkPolicies) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.GlobalNetworkPolicy, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindGlobalNetworkPolicy, noNamespace, convertPolicyNameForStorage(name))
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.GlobalNetworkPolicy), err
	}
	return nil, err
}

// Get takes name of the GlobalNetworkPolicy, and returns the corresponding GlobalNetworkPolicy object,
// and an error if there is any.
func (r globalNetworkPolicies) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.GlobalNetworkPolicy, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindGlobalNetworkPolicy, noNamespace, convertPolicyNameForStorage(name))
	if out != nil {
		// Remove the prefix out of the returned policy name.
		out.GetObjectMeta().SetName(convertPolicyNameFromStorage(out.GetObjectMeta().GetName()))
		return out.(*apiv3.GlobalNetworkPolicy), err
	}
	return nil, err
}

// List returns the list of GlobalNetworkPolicy objects that match the supplied options.
func (r globalNetworkPolicies) List(ctx context.Context, opts options.ListOptions) (*apiv3.GlobalNetworkPolicyList, error) {
	res := &apiv3.GlobalNetworkPolicyList{}
	// Add the name prefix if name is provided
	if opts.Name != "" {
		opts.Name = convertPolicyNameForStorage(opts.Name)
	}

	if err := r.client.resources.List(ctx, opts, apiv3.KindGlobalNetworkPolicy, apiv3.KindGlobalNetworkPolicyList, res); err != nil {
		return nil, err
	}

	// Remove the prefix off of each policy name
	for i, _ := range res.Items {
		name := res.Items[i].GetObjectMeta().GetName()
		res.Items[i].GetObjectMeta().SetName(convertPolicyNameFromStorage(name))
	}

	return res, nil
}

// Watch returns a watch.Interface that watches the globalNetworkPolicies that match the
// supplied options.
func (r globalNetworkPolicies) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	// Add the name prefix if name is provided
	if opts.Name != "" {
		opts.Name = convertPolicyNameForStorage(opts.Name)
	}

	return r.client.resources.Watch(ctx, opts, apiv3.KindGlobalNetworkPolicy, &policyConverter{})
}

func defaultPolicyTypesField(ingressRules, egressRules []apiv3.Rule, types *[]apiv3.PolicyType) {
	if len(*types) == 0 {
		// Default the Types field according to what inbound and outbound rules are present
		// in the policy.
		if len(egressRules) == 0 {
			// Policy has no egress rules, so apply this policy to ingress only.  (Note:
			// intentionally including the case where the policy also has no ingress
			// rules.)
			*types = []apiv3.PolicyType{apiv3.PolicyTypeIngress}
		} else if len(ingressRules) == 0 {
			// Policy has egress rules but no ingress rules, so apply this policy to
			// egress only.
			*types = []apiv3.PolicyType{apiv3.PolicyTypeEgress}
		} else {
			// Policy has both ingress and egress rules, so apply this policy to both
			// ingress and egress.
			*types = []apiv3.PolicyType{apiv3.PolicyTypeIngress, apiv3.PolicyTypeEgress}
		}
	}
}

func convertPolicyNameForStorage(name string) string {
	// Do nothing on names prefixed with "knp."
	if strings.HasPrefix(name, "knp.") {
		return name
	}
	// Similarly for "ossg."
	if strings.HasPrefix(name, "ossg.") {
		return name
	}
	return "default." + name
}

func convertPolicyNameFromStorage(name string) string {
	// Do nothing on names prefixed with "knp."
	if strings.HasPrefix(name, "knp.") {
		return name
	}
	// Similarly for "ossg."
	if strings.HasPrefix(name, "ossg.") {
		return name
	}
	parts := strings.SplitN(name, ".", 2)
	return parts[len(parts)-1]
}

type policyConverter struct{}

func (pc *policyConverter) Convert(r resource) resource {
	r.GetObjectMeta().SetName(convertPolicyNameFromStorage(r.GetObjectMeta().GetName()))
	return r
}

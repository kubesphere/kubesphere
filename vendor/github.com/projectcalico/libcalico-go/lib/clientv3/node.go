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

	"fmt"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/errors"
	"github.com/projectcalico/libcalico-go/lib/names"
	"github.com/projectcalico/libcalico-go/lib/net"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
	"github.com/projectcalico/libcalico-go/lib/options"
	validator "github.com/projectcalico/libcalico-go/lib/validator/v3"
	"github.com/projectcalico/libcalico-go/lib/watch"
	log "github.com/sirupsen/logrus"
)

// NodeInterface has methods to work with Node resources.
type NodeInterface interface {
	Create(ctx context.Context, res *apiv3.Node, opts options.SetOptions) (*apiv3.Node, error)
	Update(ctx context.Context, res *apiv3.Node, opts options.SetOptions) (*apiv3.Node, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.Node, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.Node, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.NodeList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// nodes implements NodeInterface
type nodes struct {
	client client
}

// Create takes the representation of a Node and creates it.  Returns the stored
// representation of the Node, and an error, if there is any.
func (r nodes) Create(ctx context.Context, res *apiv3.Node, opts options.SetOptions) (*apiv3.Node, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	// For host-protection only clusters, we instruct the user to create a Node as the first
	// operation.  Piggy-back the datastore initialisation on that to ensure the Ready flag gets
	// set.  Since we're likely being called from calicoctl, we don't know the Calico version.
	err := r.client.EnsureInitialized(ctx, "", "")
	if err != nil {
		return nil, err
	}
	out, err := r.client.resources.Create(ctx, opts, apiv3.KindNode, res)
	if out != nil {
		return out.(*apiv3.Node), err
	}
	return nil, err
}

// Update takes the representation of a Node and updates it. Returns the stored
// representation of the Node, and an error, if there is any.
func (r nodes) Update(ctx context.Context, res *apiv3.Node, opts options.SetOptions) (*apiv3.Node, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindNode, res)
	if out != nil {
		return out.(*apiv3.Node), err
	}
	return nil, err
}

// Delete takes name of the Node and deletes it. Returns an error if one occurs.
func (r nodes) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.Node, error) {
	pname, err := names.WorkloadEndpointIdentifiers{Node: name}.CalculateWorkloadEndpointName(true)
	if err != nil {
		return nil, err
	}

	// Get all weps belonging to the node
	weps, err := r.client.WorkloadEndpoints().List(ctx, options.ListOptions{
		Prefix: true,
		Name:   pname,
	})
	if err != nil {
		return nil, err
	}

	// Collate all IPs across all endpoints, and then release those IPs.
	ips := []net.IP{}
	for _, wep := range weps.Items {
		// The prefix match is unfortunately not a perfect match on the Node (since it is theoretically possible for
		// another node to match the prefix (e.g. a node name of the format <thisnode>-foobar would also match a prefix
		// search of the node <thisnode>). Therefore, we will also need to check that the Spec.Node field matches the Node.
		if wep.Spec.Node != name {
			continue
		}
		for _, ip := range wep.Spec.IPNetworks {
			ipAddr, _, err := cnet.ParseCIDROrIP(ip)
			if err == nil {
				ips = append(ips, *ipAddr)
			} else {
				// Validation for wep insists upon CIDR, so we should always succeed
				log.WithError(err).Warnf("Failed to parse CIDR: %s", ip)
			}
		}
	}

	// Add in tunnel addresses if they exist for the node.
	if n, err := r.client.Nodes().Get(ctx, name, options.GetOptions{}); err != nil {
		if _, ok := err.(errors.ErrorResourceDoesNotExist); !ok {
			return nil, err
		}
		// Resource does not exist, carry on and clean up as much as we can.
	} else {
		if n.Spec.BGP != nil && n.Spec.BGP.IPv4IPIPTunnelAddr != "" {
			ipAddr, _, err := cnet.ParseCIDROrIP(n.Spec.BGP.IPv4IPIPTunnelAddr)
			if err == nil {
				ips = append(ips, *ipAddr)
			} else {
				log.WithError(err).Warnf("Failed to parse IPIP tunnel address CIDR: %s", n.Spec.BGP.IPv4IPIPTunnelAddr)
			}
		}
		if n.Spec.IPv4VXLANTunnelAddr != "" {
			ipAddr, _, err := cnet.ParseCIDROrIP(n.Spec.IPv4VXLANTunnelAddr)
			if err == nil {
				ips = append(ips, *ipAddr)
			} else {
				log.WithError(err).Warnf("Failed to parse VXLAN tunnel address CIDR: %s", n.Spec.IPv4VXLANTunnelAddr)
			}
		}
	}

	_, err = r.client.IPAM().ReleaseIPs(context.Background(), ips)
	switch err.(type) {
	case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
	default:
		return nil, err
	}

	// Delete the weps.
	for _, wep := range weps.Items {
		if wep.Spec.Node != name {
			continue
		}

		_, err = r.client.WorkloadEndpoints().Delete(ctx, wep.Namespace, wep.Name, options.DeleteOptions{})
		switch err.(type) {
		case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
		default:
			return nil, err
		}
	}

	// Remove the node from the IPAM data if it exists.
	err = r.client.IPAM().RemoveIPAMHost(ctx, name)
	switch err.(type) {
	case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
	default:
		return nil, err
	}

	// Remove BGPPeers.
	bgpPeers, err := r.client.BGPPeers().List(ctx, options.ListOptions{})
	if err != nil {
		return nil, err
	}
	for _, peer := range bgpPeers.Items {
		if peer.Spec.Node != name {
			continue
		}
		_, err = r.client.BGPPeers().Delete(ctx, peer.Name, options.DeleteOptions{})
		switch err.(type) {
		case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
		default:
			return nil, err
		}
	}

	// Delete felix configuration
	nodeConfName := fmt.Sprintf("node.%s", name)
	_, err = r.client.FelixConfigurations().Delete(ctx, nodeConfName, options.DeleteOptions{})
	switch err.(type) {
	case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
	default:
		return nil, err
	}

	// Delete bgp configuration
	_, err = r.client.BGPConfigurations().Delete(ctx, nodeConfName, options.DeleteOptions{})
	switch err.(type) {
	case nil, errors.ErrorResourceDoesNotExist, errors.ErrorOperationNotSupported:
	default:
		return nil, err
	}

	// Delete the node.
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindNode, noNamespace, name)
	if out != nil {
		return out.(*apiv3.Node), err
	}
	return nil, err
}

// Get takes name of the Node, and returns the corresponding Node object,
// and an error if there is any.
func (r nodes) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.Node, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindNode, noNamespace, name)
	if out != nil {
		return out.(*apiv3.Node), err
	}
	return nil, err
}

// List returns the list of Node objects that match the supplied options.
func (r nodes) List(ctx context.Context, opts options.ListOptions) (*apiv3.NodeList, error) {
	res := &apiv3.NodeList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindNode, apiv3.KindNodeList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the Nodes that match the
// supplied options.
func (r nodes) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindNode, nil)
}

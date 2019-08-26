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

// BGPPeerInterface has methods to work with BGPPeer resources.
type BGPPeerInterface interface {
	Create(ctx context.Context, res *apiv3.BGPPeer, opts options.SetOptions) (*apiv3.BGPPeer, error)
	Update(ctx context.Context, res *apiv3.BGPPeer, opts options.SetOptions) (*apiv3.BGPPeer, error)
	Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.BGPPeer, error)
	Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.BGPPeer, error)
	List(ctx context.Context, opts options.ListOptions) (*apiv3.BGPPeerList, error)
	Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error)
}

// bgpPeers implements BGPPeerInterface
type bgpPeers struct {
	client client
}

// Create takes the representation of a BGPPeer and creates it.  Returns the stored
// representation of the BGPPeer, and an error, if there is any.
func (r bgpPeers) Create(ctx context.Context, res *apiv3.BGPPeer, opts options.SetOptions) (*apiv3.BGPPeer, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Create(ctx, opts, apiv3.KindBGPPeer, res)
	if out != nil {
		return out.(*apiv3.BGPPeer), err
	}
	return nil, err
}

// Update takes the representation of a BGPPeer and updates it. Returns the stored
// representation of the BGPPeer, and an error, if there is any.
func (r bgpPeers) Update(ctx context.Context, res *apiv3.BGPPeer, opts options.SetOptions) (*apiv3.BGPPeer, error) {
	if err := validator.Validate(res); err != nil {
		return nil, err
	}

	out, err := r.client.resources.Update(ctx, opts, apiv3.KindBGPPeer, res)
	if out != nil {
		return out.(*apiv3.BGPPeer), err
	}
	return nil, err
}

// Delete takes name of the BGPPeer and deletes it. Returns an error if one occurs.
func (r bgpPeers) Delete(ctx context.Context, name string, opts options.DeleteOptions) (*apiv3.BGPPeer, error) {
	out, err := r.client.resources.Delete(ctx, opts, apiv3.KindBGPPeer, noNamespace, name)
	if out != nil {
		return out.(*apiv3.BGPPeer), err
	}
	return nil, err
}

// Get takes name of the BGPPeer, and returns the corresponding BGPPeer object,
// and an error if there is any.
func (r bgpPeers) Get(ctx context.Context, name string, opts options.GetOptions) (*apiv3.BGPPeer, error) {
	out, err := r.client.resources.Get(ctx, opts, apiv3.KindBGPPeer, noNamespace, name)
	if out != nil {
		return out.(*apiv3.BGPPeer), err
	}
	return nil, err
}

// List returns the list of BGPPeer objects that match the supplied options.
func (r bgpPeers) List(ctx context.Context, opts options.ListOptions) (*apiv3.BGPPeerList, error) {
	res := &apiv3.BGPPeerList{}
	if err := r.client.resources.List(ctx, opts, apiv3.KindBGPPeer, apiv3.KindBGPPeerList, res); err != nil {
		return nil, err
	}
	return res, nil
}

// Watch returns a watch.Interface that watches the BGPPeers that match the
// supplied options.
func (r bgpPeers) Watch(ctx context.Context, opts options.ListOptions) (watch.Interface, error) {
	return r.client.resources.Watch(ctx, opts, apiv3.KindBGPPeer, nil)
}

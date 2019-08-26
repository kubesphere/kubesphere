// Copyright (c) 2016-2017 Tigera, Inc. All rights reserved.

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

package resources

import (
	"context"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"

	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

func NewPodCIDRBlockAffinityClient(c *kubernetes.Clientset) K8sResourceClient {
	return &podCIDRBlockClient{
		clientSet: c,
	}
}

// podCIDRBlockClient implements the api.Client interface for block affinities using Kubernetes pod CIDR
// allocations as the backing store. For use with host-local IPAM. For the Calico IPAM
// implementation, see ipam_block.go.
type podCIDRBlockClient struct {
	clientSet *kubernetes.Clientset
}

func (c *podCIDRBlockClient) Create(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Warn("Operation Create is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: kvp.Key,
		Operation:  "Create",
	}
}

func (c *podCIDRBlockClient) Update(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Warn("Operation Update is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: kvp.Key,
		Operation:  "Create",
	}
}

func (c *podCIDRBlockClient) DeleteKVP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	log.Warn("Operation DeleteKVP is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: kvp.Key,
		Operation:  "DeleteKVP",
	}
}

func (c *podCIDRBlockClient) Delete(ctx context.Context, key model.Key, revision string, uid *types.UID) (*model.KVPair, error) {
	log.Warn("Operation Delete is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: key,
		Operation:  "Delete",
	}
}

func (c *podCIDRBlockClient) Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error) {
	log.Warn("Operation Get is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: key,
		Operation:  "Get",
	}
}

func (c *podCIDRBlockClient) Watch(ctx context.Context, list model.ListInterface, revision string) (api.WatchInterface, error) {
	log.Debug("Operation Watch is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: list,
		Operation:  "Watch",
	}
}

func (c *podCIDRBlockClient) List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error) {
	log.Debug("Received List request on block affinities (using host-local IPAM)")
	bl := list.(model.BlockAffinityListOptions)
	kvpl := &model.KVPairList{
		KVPairs:  []*model.KVPair{},
		Revision: revision,
	}

	// If a host is specified, then do an exact lookup (ip version should not be expected in the query)
	if bl.Host != "" && bl.IPVersion == 0 {
		// Get the node settings, we use the nodes PodCIDR as the only node affinity block.
		node, err := c.clientSet.CoreV1().Nodes().Get(bl.Host, metav1.GetOptions{ResourceVersion: revision})
		if err != nil {
			err = K8sErrorToCalico(err, list)
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
				return nil, err
			}
			return kvpl, nil
		}

		// Return no results if the pod CIDR is not assigned.
		podcidr := node.Spec.PodCIDR
		if len(podcidr) == 0 {
			return kvpl, nil
		}

		_, cidr, err := cnet.ParseCIDR(podcidr)
		if err != nil {
			return nil, err
		}
		kvpl.Revision = node.ResourceVersion
		kvpl.KVPairs = append(kvpl.KVPairs, &model.KVPair{
			Key: model.BlockAffinityKey{
				CIDR: *cidr,
				Host: bl.Host,
			},
			Value:    &model.BlockAffinity{State: model.StateConfirmed},
			Revision: node.ResourceVersion,
		})

		return kvpl, nil
	}

	// When host is not specified...
	if bl.IPVersion == 0 {
		// Get the node settings, we use the nodes PodCIDR as the only node affinity block.
		nodeList, err := c.clientSet.CoreV1().Nodes().List(metav1.ListOptions{ResourceVersion: revision})
		if err != nil {
			err = K8sErrorToCalico(err, list)
			if _, ok := err.(cerrors.ErrorResourceDoesNotExist); !ok {
				return nil, err
			}
			return kvpl, nil
		}

		kvpl.Revision = nodeList.ResourceVersion
		for _, node := range nodeList.Items {
			// Return no results if the pod CIDR is not assigned.
			podcidr := node.Spec.PodCIDR
			if len(podcidr) == 0 {
				continue
			}

			_, cidr, err := cnet.ParseCIDR(podcidr)
			if err != nil {
				return nil, err
			}
			kvpl.KVPairs = append(kvpl.KVPairs, &model.KVPair{
				Key: model.BlockAffinityKey{
					CIDR: *cidr,
					Host: node.Name,
				},
				Value:    &model.BlockAffinity{State: model.StateConfirmed},
				Revision: node.ResourceVersion,
			})
		}
		return kvpl, nil
	}

	// Currently querying the affinity block is only used by the BGP syncer *and* we always
	// query for a specific Node, so for now fail List requests for all nodes.
	log.Warn("Operation List (all nodes or all IP versions) is not supported on block affinities when using host-local IPAM")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: list,
		Operation:  "List",
	}
}

func (c *podCIDRBlockClient) EnsureInitialized() error {
	return nil
}

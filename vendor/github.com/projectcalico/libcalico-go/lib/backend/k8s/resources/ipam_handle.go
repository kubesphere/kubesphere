// Copyright (c) 2019 Tigera, Inc. All rights reserved.

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
	"fmt"
	"reflect"
	"strings"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	cerrors "github.com/projectcalico/libcalico-go/lib/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	IPAMHandleResourceName = "IPAMHandles"
	IPAMHandleCRDName      = "ipamhandles.crd.projectcalico.org"
)

func NewIPAMHandleClient(c *kubernetes.Clientset, r *rest.RESTClient) K8sResourceClient {
	// Create a resource client which manages k8s CRDs.
	rc := customK8sResourceClient{
		clientSet:       c,
		restClient:      r,
		name:            IPAMHandleCRDName,
		resource:        IPAMHandleResourceName,
		description:     "Calico IPAM handles",
		k8sResourceType: reflect.TypeOf(apiv3.IPAMHandle{}),
		k8sResourceTypeMeta: metav1.TypeMeta{
			Kind:       apiv3.KindIPAMHandle,
			APIVersion: apiv3.GroupVersionCurrent,
		},
		k8sListType:  reflect.TypeOf(apiv3.IPAMHandleList{}),
		resourceKind: apiv3.KindIPAMHandle,
	}

	return &ipamHandleClient{rc: rc}
}

// affinityHandleClient implements the api.Client interface for IPAMHandle objects. It
// handles the translation between v1 objects understood by the IPAM codebase in lib/ipam,
// and the CRDs which are used to actually store the data in the Kubernetes API.
// It uses a customK8sResourceClient under the covers to perform CRUD operations on
// kubernetes CRDs.
type ipamHandleClient struct {
	rc customK8sResourceClient
}

func (c ipamHandleClient) toV1(kvpv3 *model.KVPair) *model.KVPair {
	handle := kvpv3.Value.(*apiv3.IPAMHandle).Spec.HandleID
	block := kvpv3.Value.(*apiv3.IPAMHandle).Spec.Block
	return &model.KVPair{
		Key: model.IPAMHandleKey{
			HandleID: handle,
		},
		Value: &model.IPAMHandle{
			HandleID: handle,
			Block:    block,
		},
		Revision: kvpv3.Revision,
	}
}

func (c ipamHandleClient) parseKey(k model.Key) string {
	return strings.ToLower(k.(model.IPAMHandleKey).HandleID)
}

func (c ipamHandleClient) toV3(kvpv1 *model.KVPair) *model.KVPair {
	name := c.parseKey(kvpv1.Key)
	handle := kvpv1.Key.(model.IPAMHandleKey).HandleID
	block := kvpv1.Value.(*model.IPAMHandle).Block
	return &model.KVPair{
		Key: model.ResourceKey{
			Name: name,
			Kind: apiv3.KindIPAMHandle,
		},
		Value: &apiv3.IPAMHandle{
			TypeMeta: metav1.TypeMeta{
				Kind:       apiv3.KindIPAMHandle,
				APIVersion: "crd.projectcalico.org/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				ResourceVersion: kvpv1.Revision,
			},
			Spec: apiv3.IPAMHandleSpec{
				HandleID: handle,
				Block:    block,
			},
		},
		Revision: kvpv1.Revision,
	}
}

func (c *ipamHandleClient) Create(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	nkvp := c.toV3(kvp)
	kvp, err := c.rc.Create(ctx, nkvp)
	if err != nil {
		return nil, err
	}
	return c.toV1(kvp), nil
}

func (c *ipamHandleClient) Update(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	nkvp := c.toV3(kvp)
	kvp, err := c.rc.Update(ctx, nkvp)
	if err != nil {
		return nil, err
	}
	return c.toV1(kvp), nil
}

func (c *ipamHandleClient) DeleteKVP(ctx context.Context, kvp *model.KVPair) (*model.KVPair, error) {
	// We need to mark as deleted first, since the Kubernetes API doesn't support
	// compare-and-delete. This update operation allows us to eliminate races with other clients.
	name := c.parseKey(kvp.Key)
	kvp.Value.(*model.IPAMHandle).Deleted = true
	v1kvp, err := c.Update(ctx, kvp)
	if err != nil {
		return nil, err
	}

	// Now actually delete the object.
	k := model.ResourceKey{Name: name, Kind: apiv3.KindIPAMHandle}
	kvp, err = c.rc.Delete(ctx, k, v1kvp.Revision, kvp.UID)
	if err != nil {
		return nil, err
	}
	return c.toV1(kvp), nil
}

func (c *ipamHandleClient) Delete(ctx context.Context, key model.Key, revision string, uid *types.UID) (*model.KVPair, error) {
	// Delete should not be used for handles, since we need the object UID for correctness.
	log.Warn("Operation Delete is not supported on IPAMHandle type")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: key,
		Operation:  "Delete",
	}
}

func (c *ipamHandleClient) Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error) {
	name := c.parseKey(key)
	k := model.ResourceKey{Name: name, Kind: apiv3.KindIPAMHandle}
	kvp, err := c.rc.Get(ctx, k, revision)
	if err != nil {
		return nil, err
	}

	// Convert it to v1.
	v1kvp := c.toV1(kvp)

	// If this object has been marked as deleted, then we need to clean it up and
	// return not found.
	if v1kvp.Value.(*model.IPAMHandle).Deleted {
		if _, err := c.DeleteKVP(ctx, v1kvp); err != nil {
			return nil, err
		}
		return nil, cerrors.ErrorResourceDoesNotExist{fmt.Errorf("Resource was deleted"), key}
	}

	return v1kvp, nil
}

func (c *ipamHandleClient) List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error) {
	l := model.ResourceListOptions{Kind: apiv3.KindIPAMHandle}
	v3list, err := c.rc.List(ctx, l, revision)
	if err != nil {
		return nil, err
	}

	kvpl := &model.KVPairList{KVPairs: []*model.KVPair{}}
	for _, i := range v3list.KVPairs {
		v1kvp := c.toV1(i)
		kvpl.KVPairs = append(kvpl.KVPairs, v1kvp)
	}
	return kvpl, nil
}

func (c *ipamHandleClient) Watch(ctx context.Context, list model.ListInterface, revision string) (api.WatchInterface, error) {
	log.Warn("Operation Watch is not supported on IPAMHandle type")
	return nil, cerrors.ErrorOperationNotSupported{
		Identifier: list,
		Operation:  "Watch",
	}
}

func (c *ipamHandleClient) EnsureInitialized() error {
	return nil
}

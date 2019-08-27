// Copyright (c) 2017-2019 Tigera, Inc. All rights reserved.

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

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/projectcalico/libcalico-go/lib/backend/api"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
)

// K8sResourceClient is the interface to the k8s datastore for CRUD operations
// on an individual resource (one for each of the *model* types supported by
// the K8s backend).
//
// Defining a separate client interface from api.Client allows the k8s-specific
// client to diverge.
type K8sResourceClient interface {
	// Create creates the object specified in the KVPair, which must not
	// already exist. On success, returns a KVPair for the object with
	// revision  information filled-in.
	Create(ctx context.Context, object *model.KVPair) (*model.KVPair, error)

	// Update modifies the existing object specified in the KVPair.
	// On success, returns a KVPair for the object with revision
	// information filled-in.  If the input KVPair has revision
	// information then the update only succeeds if the revision is still
	// current.
	Update(ctx context.Context, object *model.KVPair) (*model.KVPair, error)

	// Delete removes the object specified by the Key.  If the call
	// contains revision information, the delete only succeeds if the
	// revision is still current.
	Delete(ctx context.Context, key model.Key, revision string, uid *types.UID) (*model.KVPair, error)

	// DeleteKVP removes the object specified by the KVPair.  If the KVPair
	// contains revision information, the delete only succeeds if the
	// revision is still current.
	DeleteKVP(ctx context.Context, object *model.KVPair) (*model.KVPair, error)

	// Get returns the object identified by the given key as a KVPair with
	// revision information.
	Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error)

	// List returns a slice of KVPairs matching the input list options.
	// list should be passed one of the model.<Type>ListOptions structs.
	// Non-zero fields in the struct are used as filters.
	List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error)

	// Watch returns a WatchInterface used for watching resources matching the
	// input list options.
	Watch(ctx context.Context, list model.ListInterface, revision string) (api.WatchInterface, error)

	// EnsureInitialized ensures that the backend is initialized
	// any ready to be used.
	EnsureInitialized() error
}

// K8sNodeResourceClient extends the K8sResourceClient to add a helper method to
// extract resources from the supplied K8s Node.  This convenience interface is
// expected to be removed in a future libcalico-go release.
type K8sNodeResourceClient interface {
	K8sResourceClient
	ExtractResourcesFromNode(node *apiv1.Node) ([]*model.KVPair, error)
}

// Copyright (c) 2016-2018 Tigera, Inc. All rights reserved.
//
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

package api

import (
	"fmt"
	"sync"

	"context"

	"github.com/projectcalico/libcalico-go/lib/backend/model"
)

// SyncStatus represents the overall state of the datastore.
// When the status changes, the Syncer calls OnStatusUpdated() on its callback.
type SyncStatus uint8

const (
	// WaitForDatastore means the Syncer is waiting to connect to the datastore.
	// (Or, it is waiting for the data in the datastore to be ready to use.)
	WaitForDatastore SyncStatus = iota
	// ResyncInProgress means the Syncer is resyncing with the datastore.
	// During the first resync, the Syncer sends updates for all keys that
	// exist in the datastore as well as any updates that occur
	// concurrently.
	ResyncInProgress
	// InSync means the Syncer has now sent all the existing keys in the
	// datastore and the user of hte API has the full picture.
	InSync
)

func (s SyncStatus) String() string {
	switch s {
	case WaitForDatastore:
		return "wait-for-ready"
	case InSync:
		return "in-sync"
	case ResyncInProgress:
		return "resync"
	default:
		return fmt.Sprintf("Unknown<%v>", uint8(s))
	}
}

// Client is the interface to the backend datastore.  It makes heavy use of the
// KVPair struct, which contains a key and (optional) value drawn from the
// backend/model package along with opaque revision information that the
// datastore uses to enforce consistency.
type Client interface {
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

	// Apply updates or creates the object specified in the KVPair.
	// On success, returns a KVPair for the object with revision
	// information filled-in.  Revision information is ignored on an Apply.
	Apply(ctx context.Context, object *model.KVPair) (*model.KVPair, error)

	// Delete removes the object specified by the key.  If the call
	// contains revision information, the delete only succeeds if the
	// revision is still current.
	//
	// Some keys are hierarchical, and Delete is a recursive operation.
	//
	// Any objects that were implicitly added by a Create operation should
	// also be removed when deleting the objects that implicitly created it.
	// For example, deleting the last WorkloadEndpoint in a Workload will
	// also remove the Workload.
	Delete(ctx context.Context, key model.Key, revision string) (*model.KVPair, error)

	// DeleteKVP removes the object specified by the KVPair.  If the KVPair
	// contains revision information, the delete only succeeds if the
	// revision is still current.
	//
	// Some keys are hierarchical, and Delete is a recursive operation.
	//
	// Any objects that were implicitly added by a Create operation should
	// also be removed when deleting the objects that implicitly created it.
	// For example, deleting the last WorkloadEndpoint in a Workload will
	// also remove the Workload.
	DeleteKVP(ctx context.Context, object *model.KVPair) (*model.KVPair, error)

	// Get returns the object identified by the given key as a KVPair with
	// revision information.
	Get(ctx context.Context, key model.Key, revision string) (*model.KVPair, error)

	// List returns a slice of KVPairs matching the input list options.
	// list should be passed one of the model.<Type>ListOptions structs.
	// Non-zero fields in the struct are used as filters.
	List(ctx context.Context, list model.ListInterface, revision string) (*model.KVPairList, error)

	// Watch returns a WatchInterface used for watching a resources matching the
	// input list options.
	Watch(ctx context.Context, list model.ListInterface, revision string) (WatchInterface, error)

	// EnsureInitialized ensures that the backend is initialized
	// any ready to be used.
	EnsureInitialized() error

	// Clean removes Calico data from the backend datastore.  Used for test purposes.
	Clean() error

	// Close the client.
	//Close()
}

type Syncer interface {
	// Starts the Syncer.  May start a background goroutine.
	Start()
	// Stops the Syncer.  Stops all background goroutines. Any cached updates that the syncer knows about
	// are emitted as delete events.
	Stop()
}

type SyncerCallbacks interface {
	// OnStatusUpdated is called when the status of the sync status of the
	// datastore changes.
	OnStatusUpdated(status SyncStatus)

	// OnUpdates is called when the Syncer has one or more updates to report.
	// Updates consist of typed key-value pairs.  The keys are drawn from the
	// backend.model package.  The values are either nil, to indicate a
	// deletion (or failure to parse a value), or a pointer to a value of
	// the associated value type.
	//
	// When a recursive delete is made, deleting many leaf keys, the Syncer
	// generates deletion updates for all the leaf keys.
	OnUpdates(updates []Update)
}

// SyncerParseFailCallbacks is an optional interface that can be implemented
// by a Syncer callback.  Datastores that support it can report a failure to
// parse a particular key or value.
type SyncerParseFailCallbacks interface {
	ParseFailed(rawKey string, rawValue string)
}

// Update from the Syncer.  A KV pair plus extra metadata.
type Update struct {
	model.KVPair
	UpdateType UpdateType
}

type UpdateType uint8

const (
	UpdateTypeKVUnknown UpdateType = iota
	UpdateTypeKVNew
	UpdateTypeKVUpdated
	UpdateTypeKVDeleted
)

// Interface can be implemented by anything that knows how to watch and report changes.
type WatchInterface interface {
	// Stops watching. Will close the channel returned by ResultChan(). Releases
	// any resources used by the watch.
	Stop()

	// Returns a chan which will receive all the events.  This channel is closed when:
	// -  Stop() is called, or
	// -  A error of type errors.ErrorWatchTerminated is received.
	// In both cases the watcher will be cleaned up, and the client should stop receiving
	// from this channel.
	ResultChan() <-chan WatchEvent

	// HasTerminated returns true if the watcher has terminated and released all
	// resources.  This is used for test purposes.
	HasTerminated() bool
}

// WatchEventType defines the possible types of events.
type WatchEventType string

const (
	WatchAdded    WatchEventType = "ADDED"
	WatchModified WatchEventType = "MODIFIED"
	WatchDeleted  WatchEventType = "DELETED"
	WatchError    WatchEventType = "ERROR"
)

// Event represents a single event to a watched resource.
type WatchEvent struct {
	Type WatchEventType

	// Old is:
	// * If Type is Added or Error: nil
	// * If Type is Modified or Deleted: the previous state of the object
	// New is:
	//  * If Type is Added or Modified: the new state of the object.
	//  * If Type is Deleted or Error: nil
	Old *model.KVPair
	New *model.KVPair

	// The error, if EventType is Error.
	Error error
}

// FakeWatcher is inspired by apimachinery (watch) FakeWatcher
type FakeWatcher struct {
	result  chan WatchEvent
	Stopped bool
	sync.Mutex
}

// NewFake constructs a FakeWatcher
func NewFake() *FakeWatcher {
	return &FakeWatcher{
		result: make(chan WatchEvent),
	}
}

// Stop implements WatchInterface
func (f *FakeWatcher) Stop() {
	f.Lock()
	defer f.Unlock()
	if !f.Stopped {
		close(f.result)
		f.Stopped = true
	}
}

// ResultChan implements WatchInterface
func (f *FakeWatcher) ResultChan() <-chan WatchEvent {
	return f.result
}

// HasTerminated implements WatchInterface
func (f *FakeWatcher) HasTerminated() bool {
	return false
}

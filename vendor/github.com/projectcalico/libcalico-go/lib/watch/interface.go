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

package watch

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// Interface can be implemented by anything that knows how to watch and report changes.
type Interface interface {
	// Stops watching. Will close the channel returned by ResultChan(). Releases
	// any resources used by the watch.
	Stop()

	// Returns a chan which will receive all the events. If an error occurs
	// or Stop() is called, this channel will be closed, in which case the
	// watch should be completely cleaned up.
	ResultChan() <-chan Event
}

// EventType defines the possible types of events.
type EventType string

const (
	// Event type:
	// Added
	// * a new Object has been added.  If the Watcher does not have a specific
	//   ResourceVersion to watch from, existing entries will first be listed
	//   and propagated as "Added" events.
	// Modified
	// * an Object has been modified.
	// Deleted
	// * an Object has been deleted
	// Error
	// * an error has occurred.  If the error is terminating, the results channel
	//   will be closed.
	Added    EventType = "ADDED"
	Modified EventType = "MODIFIED"
	Deleted  EventType = "DELETED"
	Error    EventType = "ERROR"

	DefaultChanSize int32 = 100
)

// Event represents a single event to a watched resource.
type Event struct {
	Type EventType

	// Previous is:
	// * If Type is Added, Error or Synced: nil
	// * If Type is Modified or Deleted: the previous state of the object
	// Object is:
	//  * If Type is Added or Modified: the new state of the object.
	//  * If Type is Deleted, Error or Synced: nil
	Previous runtime.Object
	Object   runtime.Object

	// The error, if EventType is Error.
	Error error
}

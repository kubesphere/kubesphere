/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

import (
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Signal defines a signal that can trigger eviction of pods on a node.
type Signal string

const (
	// SignalMemoryAvailable is memory available (i.e. capacity - workingSet), in bytes.
	SignalMemoryAvailable Signal = "memory.available"
	// SignalNodeFsAvailable is amount of storage available on filesystem that kubelet uses for volumes, daemon logs, etc.
	SignalNodeFsAvailable Signal = "nodefs.available"
	// SignalNodeFsInodesFree is amount of inodes available on filesystem that kubelet uses for volumes, daemon logs, etc.
	SignalNodeFsInodesFree Signal = "nodefs.inodesFree"
	// SignalImageFsAvailable is amount of storage available on filesystem that container runtime uses for storing images and container writable layers.
	SignalImageFsAvailable Signal = "imagefs.available"
	// SignalImageFsInodesFree is amount of inodes available on filesystem that container runtime uses for storing images and container writeable layers.
	SignalImageFsInodesFree Signal = "imagefs.inodesFree"
	// SignalAllocatableMemoryAvailable is amount of memory available for pod allocation (i.e. allocatable - workingSet (of pods), in bytes.
	SignalAllocatableMemoryAvailable Signal = "allocatableMemory.available"
	// SignalAllocatableNodeFsAvailable is amount of local storage available for pod allocation
	SignalAllocatableNodeFsAvailable Signal = "allocatableNodeFs.available"
	// SignalPIDAvailable is amount of PID available for pod allocation
	SignalPIDAvailable Signal = "pid.available"
)

// ThresholdOperator is the operator used to express a Threshold.
type ThresholdOperator string

const (
	// OpLessThan is the operator that expresses a less than operator.
	OpLessThan ThresholdOperator = "LessThan"
)

// OpForSignal maps Signals to ThresholdOperators.
// Today, the only supported operator is "LessThan". This may change in the future,
// for example if "consumed" (as opposed to "available") type signals are added.
// In both cases the directionality of the threshold is implicit to the signal type
// (for a given signal, the decision to evict will be made when crossing the threshold
// from either above or below, never both). There is thus no reason to expose the
// operator in the Kubelet's public API. Instead, we internally map signal types to operators.
var OpForSignal = map[Signal]ThresholdOperator{
	SignalMemoryAvailable:            OpLessThan,
	SignalNodeFsAvailable:            OpLessThan,
	SignalNodeFsInodesFree:           OpLessThan,
	SignalImageFsAvailable:           OpLessThan,
	SignalImageFsInodesFree:          OpLessThan,
	SignalAllocatableMemoryAvailable: OpLessThan,
	SignalAllocatableNodeFsAvailable: OpLessThan,
}

// ThresholdValue is a value holder that abstracts literal versus percentage based quantity
type ThresholdValue struct {
	// The following fields are exclusive. Only the topmost non-zero field is used.

	// Quantity is a quantity associated with the signal that is evaluated against the specified operator.
	Quantity *resource.Quantity
	// Percentage represents the usage percentage over the total resource that is evaluated against the specified operator.
	Percentage float32
}

// Threshold defines a metric for when eviction should occur.
type Threshold struct {
	// Signal defines the entity that was measured.
	Signal Signal
	// Operator represents a relationship of a signal to a value.
	Operator ThresholdOperator
	// Value is the threshold the resource is evaluated against.
	Value ThresholdValue
	// GracePeriod represents the amount of time that a threshold must be met before eviction is triggered.
	GracePeriod time.Duration
	// MinReclaim represents the minimum amount of resource to reclaim if the threshold is met.
	MinReclaim *ThresholdValue
}

// GetThresholdQuantity returns the expected quantity value for a thresholdValue
func GetThresholdQuantity(value ThresholdValue, capacity *resource.Quantity) *resource.Quantity {
	if value.Quantity != nil {
		return value.Quantity.Copy()
	}
	return resource.NewQuantity(int64(float64(capacity.Value())*float64(value.Percentage)), resource.BinarySI)
}

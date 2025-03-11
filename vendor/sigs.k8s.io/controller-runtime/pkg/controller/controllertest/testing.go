/*
Copyright 2018 The Kubernetes Authors.

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

package controllertest

import (
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ runtime.Object = &ErrorType{}

// ErrorType implements runtime.Object but isn't registered in any scheme and should cause errors in tests as a result.
type ErrorType struct{}

// GetObjectKind implements runtime.Object.
func (ErrorType) GetObjectKind() schema.ObjectKind { return nil }

// DeepCopyObject implements runtime.Object.
func (ErrorType) DeepCopyObject() runtime.Object { return nil }

var _ workqueue.TypedRateLimitingInterface[reconcile.Request] = &Queue{}

// Queue implements a RateLimiting queue as a non-ratelimited queue for testing.
// This helps testing by having functions that use a RateLimiting queue synchronously add items to the queue.
type Queue = TypedQueue[reconcile.Request]

// TypedQueue implements a RateLimiting queue as a non-ratelimited queue for testing.
// This helps testing by having functions that use a RateLimiting queue synchronously add items to the queue.
type TypedQueue[request comparable] struct {
	workqueue.TypedInterface[request]
	AddedRateLimitedLock sync.Mutex
	AddedRatelimited     []any
}

// AddAfter implements RateLimitingInterface.
func (q *TypedQueue[request]) AddAfter(item request, duration time.Duration) {
	q.Add(item)
}

// AddRateLimited implements RateLimitingInterface.  TODO(community): Implement this.
func (q *TypedQueue[request]) AddRateLimited(item request) {
	q.AddedRateLimitedLock.Lock()
	q.AddedRatelimited = append(q.AddedRatelimited, item)
	q.AddedRateLimitedLock.Unlock()
	q.Add(item)
}

// Forget implements RateLimitingInterface.  TODO(community): Implement this.
func (q *TypedQueue[request]) Forget(item request) {}

// NumRequeues implements RateLimitingInterface.  TODO(community): Implement this.
func (q *TypedQueue[request]) NumRequeues(item request) int {
	return 0
}

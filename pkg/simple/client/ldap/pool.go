/*
Copyright 2019 The KubeSphere Authors.

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

package ldap

import (
	"errors"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// Pool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type Pool interface {
	// Get returns a new connection from the pool. Closing the connections puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error.
	Get() (*PoolConn, error)

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}

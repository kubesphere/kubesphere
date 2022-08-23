// SPDX-License-Identifier: Apache-2.0
//
// The OpenSearch Contributors require contributions made to
// this file be licensed under the Apache-2.0 license or a
// compatible open source license.
//
// Modifications Copyright OpenSearch Contributors. See
// GitHub history for details.

// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package opensearchtransport

import (
	"errors"
	"fmt"
	"math"
	"net/url"
	"sort"
	"sync"
	"time"
)

var (
	defaultResurrectTimeoutInitial      = 60 * time.Second
	defaultResurrectTimeoutFactorCutoff = 5
)

// Selector defines the interface for selecting connections from the pool.
//
type Selector interface {
	Select([]*Connection) (*Connection, error)
}

// ConnectionPool defines the interface for the connection pool.
//
type ConnectionPool interface {
	Next() (*Connection, error)  // Next returns the next available connection.
	OnSuccess(*Connection) error // OnSuccess reports that the connection was successful.
	OnFailure(*Connection) error // OnFailure reports that the connection failed.
	URLs() []*url.URL            // URLs returns the list of URLs of available connections.
}

// Connection represents a connection to a node.
//
type Connection struct {
	sync.Mutex

	URL       *url.URL
	IsDead    bool
	DeadSince time.Time
	Failures  int

	ID         string
	Name       string
	Roles      []string
	Attributes map[string]interface{}
}

type singleConnectionPool struct {
	connection *Connection

	metrics *metrics
}

type statusConnectionPool struct {
	sync.Mutex

	live     []*Connection // List of live connections
	dead     []*Connection // List of dead connections
	selector Selector

	metrics *metrics
}

type roundRobinSelector struct {
	sync.Mutex

	curr int // Index of the current connection
}

// NewConnectionPool creates and returns a default connection pool.
//
func NewConnectionPool(conns []*Connection, selector Selector) (ConnectionPool, error) {
	if len(conns) == 1 {
		return &singleConnectionPool{connection: conns[0]}, nil
	}
	if selector == nil {
		selector = &roundRobinSelector{curr: -1}
	}
	return &statusConnectionPool{live: conns, selector: selector}, nil
}

// Next returns the connection from pool.
//
func (cp *singleConnectionPool) Next() (*Connection, error) {
	return cp.connection, nil
}

// OnSuccess is a no-op for single connection pool.
func (cp *singleConnectionPool) OnSuccess(c *Connection) error { return nil }

// OnFailure is a no-op for single connection pool.
func (cp *singleConnectionPool) OnFailure(c *Connection) error { return nil }

// URLs returns the list of URLs of available connections.
func (cp *singleConnectionPool) URLs() []*url.URL { return []*url.URL{cp.connection.URL} }

func (cp *singleConnectionPool) connections() []*Connection { return []*Connection{cp.connection} }

// Next returns a connection from pool, or an error.
//
func (cp *statusConnectionPool) Next() (*Connection, error) {
	cp.Lock()
	defer cp.Unlock()

	// Return next live connection
	if len(cp.live) > 0 {
		return cp.selector.Select(cp.live)
	} else if len(cp.dead) > 0 {
		// No live connection is available, resurrect one of the dead ones.
		c := cp.dead[len(cp.dead)-1]
		cp.dead = cp.dead[:len(cp.dead)-1]
		c.Lock()
		defer c.Unlock()
		cp.resurrect(c, false)
		return c, nil
	}
	return nil, errors.New("no connection available")
}

// OnSuccess marks the connection as successful.
//
func (cp *statusConnectionPool) OnSuccess(c *Connection) error {
	c.Lock()
	defer c.Unlock()

	// Short-circuit for live connection
	if !c.IsDead {
		return nil
	}

	c.markAsHealthy()

	cp.Lock()
	defer cp.Unlock()
	return cp.resurrect(c, true)
}

// OnFailure marks the connection as failed.
//
func (cp *statusConnectionPool) OnFailure(c *Connection) error {
	cp.Lock()
	defer cp.Unlock()

	c.Lock()

	if c.IsDead {
		if debugLogger != nil {
			debugLogger.Logf("Already removed %s\n", c.URL)
		}
		c.Unlock()
		return nil
	}

	if debugLogger != nil {
		debugLogger.Logf("Removing %s...\n", c.URL)
	}
	c.markAsDead()
	cp.scheduleResurrect(c)
	c.Unlock()

	// Push item to dead list and sort slice by number of failures
	cp.dead = append(cp.dead, c)
	sort.Slice(cp.dead, func(i, j int) bool {
		c1 := cp.dead[i]
		c2 := cp.dead[j]
		c1.Lock()
		c2.Lock()
		defer c1.Unlock()
		defer c2.Unlock()

		res := c1.Failures > c2.Failures
		return res
	})

	// Check if connection exists in the list, return error if not.
	index := -1
	for i, conn := range cp.live {
		if conn == c {
			index = i
		}
	}
	if index < 0 {
		return errors.New("connection not in live list")
	}

	// Remove item; https://github.com/golang/go/wiki/SliceTricks
	copy(cp.live[index:], cp.live[index+1:])
	cp.live = cp.live[:len(cp.live)-1]

	return nil
}

// URLs returns the list of URLs of available connections.
//
func (cp *statusConnectionPool) URLs() []*url.URL {
	var urls []*url.URL

	cp.Lock()
	defer cp.Unlock()

	for _, c := range cp.live {
		urls = append(urls, c.URL)
	}

	return urls
}

func (cp *statusConnectionPool) connections() []*Connection {
	var conns []*Connection
	conns = append(conns, cp.live...)
	conns = append(conns, cp.dead...)
	return conns
}

// resurrect adds the connection to the list of available connections.
// When removeDead is true, it also removes it from the dead list.
// The calling code is responsible for locking.
//
func (cp *statusConnectionPool) resurrect(c *Connection, removeDead bool) error {
	if debugLogger != nil {
		debugLogger.Logf("Resurrecting %s\n", c.URL)
	}

	c.markAsLive()
	cp.live = append(cp.live, c)

	if removeDead {
		index := -1
		for i, conn := range cp.dead {
			if conn == c {
				index = i
			}
		}
		if index >= 0 {
			// Remove item; https://github.com/golang/go/wiki/SliceTricks
			copy(cp.dead[index:], cp.dead[index+1:])
			cp.dead = cp.dead[:len(cp.dead)-1]
		}
	}

	return nil
}

// scheduleResurrect schedules the connection to be resurrected.
//
func (cp *statusConnectionPool) scheduleResurrect(c *Connection) {
	factor := math.Min(float64(c.Failures-1), float64(defaultResurrectTimeoutFactorCutoff))
	timeout := time.Duration(defaultResurrectTimeoutInitial.Seconds() * math.Exp2(factor) * float64(time.Second))
	if debugLogger != nil {
		debugLogger.Logf("Resurrect %s (failures=%d, factor=%1.1f, timeout=%s) in %s\n", c.URL, c.Failures, factor, timeout, c.DeadSince.Add(timeout).Sub(time.Now().UTC()).Truncate(time.Second))
	}

	time.AfterFunc(timeout, func() {
		cp.Lock()
		defer cp.Unlock()

		c.Lock()
		defer c.Unlock()

		if !c.IsDead {
			if debugLogger != nil {
				debugLogger.Logf("Already resurrected %s\n", c.URL)
			}
			return
		}

		cp.resurrect(c, true)
	})
}

// Select returns the connection in a round-robin fashion.
//
func (s *roundRobinSelector) Select(conns []*Connection) (*Connection, error) {
	s.Lock()
	defer s.Unlock()

	s.curr = (s.curr + 1) % len(conns)
	return conns[s.curr], nil
}

// markAsDead marks the connection as dead.
//
func (c *Connection) markAsDead() {
	c.IsDead = true
	if c.DeadSince.IsZero() {
		c.DeadSince = time.Now().UTC()
	}
	c.Failures++
}

// markAsLive marks the connection as alive.
//
func (c *Connection) markAsLive() {
	c.IsDead = false
}

// markAsHealthy marks the connection as healthy.
//
func (c *Connection) markAsHealthy() {
	c.IsDead = false
	c.DeadSince = time.Time{}
	c.Failures = 0
}

// String returns a readable connection representation.
//
func (c *Connection) String() string {
	c.Lock()
	defer c.Unlock()
	return fmt.Sprintf("<%s> dead=%v failures=%d", c.URL, c.IsDead, c.Failures)
}

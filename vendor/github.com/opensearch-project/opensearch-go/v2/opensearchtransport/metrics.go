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
	"strconv"
	"strings"
	"sync"
	"time"
)

// Measurable defines the interface for transports supporting metrics.
//
type Measurable interface {
	Metrics() (Metrics, error)
}

// connectionable defines the interface for transports returning a list of connections.
//
type connectionable interface {
	connections() []*Connection
}

// Metrics represents the transport metrics.
//
type Metrics struct {
	Requests  int         `json:"requests"`
	Failures  int         `json:"failures"`
	Responses map[int]int `json:"responses"`

	Connections []fmt.Stringer `json:"connections"`
}

// ConnectionMetric represents metric information for a connection.
//
type ConnectionMetric struct {
	URL       string     `json:"url"`
	Failures  int        `json:"failures,omitempty"`
	IsDead    bool       `json:"dead,omitempty"`
	DeadSince *time.Time `json:"dead_since,omitempty"`

	Meta struct {
		ID    string   `json:"id"`
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	} `json:"meta"`
}

// metrics represents the inner state of metrics.
//
type metrics struct {
	sync.RWMutex

	requests  int
	failures  int
	responses map[int]int

	connections []*Connection
}

// Metrics returns the transport metrics.
//
func (c *Client) Metrics() (Metrics, error) {
	if c.metrics == nil {
		return Metrics{}, errors.New("transport metrics not enabled")
	}
	c.metrics.RLock()
	defer c.metrics.RUnlock()

	if lockable, ok := c.pool.(sync.Locker); ok {
		lockable.Lock()
		defer lockable.Unlock()
	}

	m := Metrics{
		Requests:  c.metrics.requests,
		Failures:  c.metrics.failures,
		Responses: c.metrics.responses,
	}

	if pool, ok := c.pool.(connectionable); ok {
		for _, c := range pool.connections() {
			c.Lock()

			cm := ConnectionMetric{
				URL:      c.URL.String(),
				IsDead:   c.IsDead,
				Failures: c.Failures,
			}

			if !c.DeadSince.IsZero() {
				cm.DeadSince = &c.DeadSince
			}

			if c.ID != "" {
				cm.Meta.ID = c.ID
			}

			if c.Name != "" {
				cm.Meta.Name = c.Name
			}

			if len(c.Roles) > 0 {
				cm.Meta.Roles = c.Roles
			}

			m.Connections = append(m.Connections, cm)
			c.Unlock()
		}
	}

	return m, nil
}

// String returns the metrics as a string.
//
func (m Metrics) String() string {
	var (
		i int
		b strings.Builder
	)
	b.WriteString("{")

	b.WriteString("Requests:")
	b.WriteString(strconv.Itoa(m.Requests))

	b.WriteString(" Failures:")
	b.WriteString(strconv.Itoa(m.Failures))

	if len(m.Responses) > 0 {
		b.WriteString(" Responses: ")
		b.WriteString("[")

		for code, num := range m.Responses {
			b.WriteString(strconv.Itoa(code))
			b.WriteString(":")
			b.WriteString(strconv.Itoa(num))
			if i+1 < len(m.Responses) {
				b.WriteString(", ")
			}
			i++
		}
		b.WriteString("]")
	}

	b.WriteString(" Connections: [")
	for i, c := range m.Connections {
		b.WriteString(c.String())
		if i+1 < len(m.Connections) {
			b.WriteString(", ")
		}
		i++
	}
	b.WriteString("]")

	b.WriteString("}")
	return b.String()
}

// String returns the connection information as a string.
//
func (cm ConnectionMetric) String() string {
	var b strings.Builder
	b.WriteString("{")
	b.WriteString(cm.URL)
	if cm.IsDead {
		fmt.Fprintf(&b, " dead=%v", cm.IsDead)
	}
	if cm.Failures > 0 {
		fmt.Fprintf(&b, " failures=%d", cm.Failures)
	}
	if cm.DeadSince != nil {
		fmt.Fprintf(&b, " dead_since=%s", cm.DeadSince.Local().Format(time.Stamp))
	}
	b.WriteString("}")
	return b.String()
}

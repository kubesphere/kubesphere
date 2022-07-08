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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
)

// Discoverable defines the interface for transports supporting node discovery.
//
type Discoverable interface {
	DiscoverNodes() error
}

// nodeInfo represents the information about node in a cluster.
//
type nodeInfo struct {
	ID         string
	Name       string
	URL        *url.URL
	Roles      []string `json:"roles"`
	Attributes map[string]interface{}
	HTTP       struct {
		PublishAddress string `json:"publish_address"`
	}
}

// DiscoverNodes reloads the client connections by fetching information from the cluster.
//
func (c *Client) DiscoverNodes() error {
	var conns []*Connection

	nodes, err := c.getNodesInfo()
	if err != nil {
		if debugLogger != nil {
			debugLogger.Logf("Error getting nodes info: %s\n", err)
		}
		return fmt.Errorf("discovery: get nodes: %s", err)
	}

	for _, node := range nodes {
		var (
			isMasterOnlyNode bool
		)

		roles := append(node.Roles[:0:0], node.Roles...)
		sort.Strings(roles)

		if len(roles) == 1 && roles[0] == "master" {
			isMasterOnlyNode = true
		}

		if debugLogger != nil {
			var skip string
			if isMasterOnlyNode {
				skip = "; [SKIP]"
			}
			debugLogger.Logf("Discovered node [%s]; %s; roles=%s%s\n", node.Name, node.URL, node.Roles, skip)
		}

		// Skip master only nodes
		// TODO: Move logic to Selector?
		if isMasterOnlyNode {
			continue
		}

		conns = append(conns, &Connection{
			URL:        node.URL,
			ID:         node.ID,
			Name:       node.Name,
			Roles:      node.Roles,
			Attributes: node.Attributes,
		})
	}

	c.Lock()
	defer c.Unlock()

	if lockable, ok := c.pool.(sync.Locker); ok {
		lockable.Lock()
		defer lockable.Unlock()
	}

	if c.poolFunc != nil {
		c.pool = c.poolFunc(conns, c.selector)
	} else {
		// TODO: Replace only live connections, leave dead scheduled for resurrect?
		c.pool, err = NewConnectionPool(conns, c.selector)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) getNodesInfo() ([]nodeInfo, error) {
	var (
		out    []nodeInfo
		scheme = c.urls[0].Scheme
	)

	req, err := http.NewRequest("GET", "/_nodes/http", nil)
	if err != nil {
		return out, err
	}

	c.Lock()
	conn, err := c.pool.Next()
	c.Unlock()
	// TODO: If no connection is returned, fallback to original URLs
	if err != nil {
		return out, err
	}

	c.setReqURL(conn.URL, req)
	c.setReqAuth(conn.URL, req)
	c.setReqUserAgent(req)

	res, err := c.transport.RoundTrip(req)
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	if res.StatusCode > 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return out, fmt.Errorf("server error: %s: %s", res.Status, body)
	}

	var env map[string]json.RawMessage
	if err := json.NewDecoder(res.Body).Decode(&env); err != nil {
		return out, err
	}

	var nodes map[string]nodeInfo
	if err := json.Unmarshal(env["nodes"], &nodes); err != nil {
		return out, err
	}

	for id, node := range nodes {
		node.ID = id
		u, err := c.getNodeURL(node, scheme)
		if err != nil {
			return out, err
		}
		node.URL = u
		out = append(out, node)
	}

	return out, nil
}

func (c *Client) getNodeURL(node nodeInfo, scheme string) (*url.URL, error) {
	var (
		host string
		port string

		addrs = strings.Split(node.HTTP.PublishAddress, "/")
		ports = strings.Split(node.HTTP.PublishAddress, ":")
	)

	if len(addrs) > 1 {
		host = addrs[0]
	} else {
		host = strings.Split(addrs[0], ":")[0]
	}
	port = ports[len(ports)-1]

	u := &url.URL{
		Scheme: scheme,
		Host:   host + ":" + port,
	}

	return u, nil
}

func (c *Client) scheduleDiscoverNodes(d time.Duration) {
	go c.DiscoverNodes()

	c.Lock()
	defer c.Unlock()
	if c.discoverNodesTimer != nil {
		c.discoverNodesTimer.Stop()
	}
	c.discoverNodesTimer = time.AfterFunc(c.discoverNodesInterval, func() {
		c.scheduleDiscoverNodes(c.discoverNodesInterval)
	})
}

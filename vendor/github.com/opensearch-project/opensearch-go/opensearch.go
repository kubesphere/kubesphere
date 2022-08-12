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

package opensearch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/signer"

	"github.com/opensearch-project/opensearch-go/internal/version"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"github.com/opensearch-project/opensearch-go/opensearchtransport"
)

var (
	reVersion *regexp.Regexp
)

func init() {
	versionPattern := `^([0-9]+)\.([0-9]+)\.([0-9]+)`
	reVersion = regexp.MustCompile(versionPattern)
}

const (
	defaultURL          = "http://localhost:9200"
	openSearch          = "opensearch"
	unsupportedProduct  = "the client noticed that the server is not a supported distribution"
	envOpenSearchURL    = "OPENSEARCH_URL"
	envElasticsearchURL = "ELASTICSEARCH_URL"
)

// Version returns the package version as a string.
//
const Version = version.Client

// Config represents the client configuration.
//
type Config struct {
	Addresses []string // A list of nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.

	Header http.Header // Global HTTP request header.

	Signer signer.Signer

	// PEM-encoded certificate authorities.
	// When set, an empty certificate pool will be created, and the certificates will be appended to it.
	// The option is only valid when the transport is not specified, or when it's http.Transport.
	CACert []byte

	RetryOnStatus        []int // List of status codes for retry. Default: 502, 503, 504.
	DisableRetry         bool  // Default: false.
	EnableRetryOnTimeout bool  // Default: false.
	MaxRetries           int   // Default: 3.

	CompressRequestBody bool // Default: false.

	DiscoverNodesOnStart  bool          // Discover nodes when initializing the client. Default: false.
	DiscoverNodesInterval time.Duration // Discover nodes periodically. Default: disabled.

	EnableMetrics     bool // Enable the metrics collection.
	EnableDebugLogger bool // Enable the debug logging.

	UseResponseCheckOnly bool

	RetryBackoff func(attempt int) time.Duration // Optional backoff duration. Default: nil.

	Transport http.RoundTripper            // The HTTP transport object.
	Logger    opensearchtransport.Logger   // The logger object.
	Selector  opensearchtransport.Selector // The selector object.

	// Optional constructor function for a custom ConnectionPool. Default: nil.
	ConnectionPoolFunc func([]*opensearchtransport.Connection, opensearchtransport.Selector) opensearchtransport.ConnectionPool
}

// Client represents the OpenSearch client.
//
type Client struct {
	*opensearchapi.API   // Embeds the API methods
	Transport            opensearchtransport.Interface
	useResponseCheckOnly bool

	productCheckMu      sync.RWMutex
	productCheckSuccess bool
}

type esVersion struct {
	Number       string `json:"number"`
	BuildFlavor  string `json:"build_flavor"`
	Distribution string `json:"distribution"`
}

type info struct {
	Version esVersion `json:"version"`
	Tagline string    `json:"tagline"`
}

// NewDefaultClient creates a new client with default options.
//
// It will use http://localhost:9200 as the default address.
//
// It will use the OPENSEARCH_URL/ELASTICSEARCH_URL environment variable, if set,
// to configure the addresses; use a comma to separate multiple URLs.
//
// It's an error to set both OPENSEARCH_URL and ELASTICSEARCH_URL.
//
func NewDefaultClient() (*Client, error) {
	return NewClient(Config{})
}

// NewClient creates a new client with configuration from cfg.
//
// It will use http://localhost:9200 as the default address.
//
// It will use the OPENSEARCH_URL/ELASTICSEARCH_URL environment variable, if set,
// to configure the addresses; use a comma to separate multiple URLs.
//
// It's an error to set both OPENSEARCH_URL and ELASTICSEARCH_URL.
//
func NewClient(cfg Config) (*Client, error) {
	var addrs []string

	if len(cfg.Addresses) == 0 {
		envAddress, err := getAddressFromEnvironment()
		if err != nil {
			return nil, err
		}
		addrs = envAddress
	} else {
		addrs = append(addrs, cfg.Addresses...)
	}

	urls, err := addrsToURLs(addrs)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %s", err)
	}

	if len(urls) == 0 {
		u, _ := url.Parse(defaultURL) // errcheck exclude
		urls = append(urls, u)
	}

	// TODO: Refactor
	if urls[0].User != nil {
		cfg.Username = urls[0].User.Username()
		pw, _ := urls[0].User.Password()
		cfg.Password = pw
	}

	tp, err := opensearchtransport.New(opensearchtransport.Config{
		URLs:     urls,
		Username: cfg.Username,
		Password: cfg.Password,

		Header: cfg.Header,
		CACert: cfg.CACert,

		Signer: cfg.Signer,

		RetryOnStatus:        cfg.RetryOnStatus,
		DisableRetry:         cfg.DisableRetry,
		EnableRetryOnTimeout: cfg.EnableRetryOnTimeout,
		MaxRetries:           cfg.MaxRetries,
		RetryBackoff:         cfg.RetryBackoff,

		CompressRequestBody: cfg.CompressRequestBody,

		EnableMetrics:     cfg.EnableMetrics,
		EnableDebugLogger: cfg.EnableDebugLogger,

		DiscoverNodesInterval: cfg.DiscoverNodesInterval,

		Transport:          cfg.Transport,
		Logger:             cfg.Logger,
		Selector:           cfg.Selector,
		ConnectionPoolFunc: cfg.ConnectionPoolFunc,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating transport: %s", err)
	}

	client := &Client{Transport: tp, useResponseCheckOnly: cfg.UseResponseCheckOnly}
	client.API = opensearchapi.New(client)

	if cfg.DiscoverNodesOnStart {
		go client.DiscoverNodes()
	}

	return client, err
}

func getAddressFromEnvironment() ([]string, error) {
	fromOpenSearchEnv := addrsFromEnvironment(envOpenSearchURL)
	fromElasticsearchEnv := addrsFromEnvironment(envElasticsearchURL)

	if len(fromElasticsearchEnv) > 0 && len(fromOpenSearchEnv) > 0 {
		return nil, fmt.Errorf("cannot create client: both %s and %s are set", envOpenSearchURL, envElasticsearchURL)
	}
	if len(fromOpenSearchEnv) > 0 {
		return fromOpenSearchEnv, nil
	}
	return fromElasticsearchEnv, nil
}

// checkCompatibleInfo validates the information given by OpenSearch
//
func checkCompatibleInfo(info info) error {
	major, _, _, err := ParseVersion(info.Version.Number)
	if err != nil {
		return err
	}
	if info.Version.Distribution == openSearch {
		return nil
	}
	if major != 7 {
		return errors.New(unsupportedProduct)
	}
	return nil
}

// ParseVersion returns an int64 representation of version.
//
func ParseVersion(version string) (int64, int64, int64, error) {
	matches := reVersion.FindStringSubmatch(version)

	if len(matches) < 4 {
		return 0, 0, 0, fmt.Errorf("")
	}
	major, _ := strconv.ParseInt(matches[1], 10, 0)
	minor, _ := strconv.ParseInt(matches[2], 10, 0)
	patch, _ := strconv.ParseInt(matches[3], 10, 0)

	return major, minor, patch, nil
}

// Perform delegates to Transport to execute a request and return a response.
//
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	if !c.useResponseCheckOnly {
		// Launch product check, request info, check header then payload.
		if err := c.doProductCheck(c.productCheck); err != nil {
			return nil, err
		}
	}

	// Retrieve the original request.
	return c.Transport.Perform(req)
}

// doProductCheck calls f if there as not been a prior successful call to doProductCheck,
// returning nil otherwise.
func (c *Client) doProductCheck(f func() error) error {
	c.productCheckMu.RLock()
	productCheckSuccess := c.productCheckSuccess
	c.productCheckMu.RUnlock()

	if productCheckSuccess {
		return nil
	}

	c.productCheckMu.Lock()
	defer c.productCheckMu.Unlock()

	if c.productCheckSuccess {
		return nil
	}

	if err := f(); err != nil {
		return err
	}

	c.productCheckSuccess = true

	return nil
}

// productCheck runs an opensearchapi.Info query to retrieve information of the current cluster
// decodes the response and decides if the cluster can be supported or not.
func (c *Client) productCheck() error {
	req := opensearchapi.InfoRequest{}
	res, err := req.Do(context.Background(), c.Transport)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		_, err = io.Copy(ioutil.Discard, res.Body)
		if err != nil {
			return err
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return nil
		case http.StatusForbidden:
			return nil
		default:
			return fmt.Errorf("cannot retrieve information from OpenSearch")
		}
	}

	var info info
	contentType := res.Header.Get("Content-Type")
	if strings.Contains(contentType, "json") {
		err = json.NewDecoder(res.Body).Decode(&info)
		if err != nil {
			return fmt.Errorf("error decoding OpenSearch informations: %s", err)
		}
	}

	if info.Version.Number != "" {
		return checkCompatibleInfo(info)
	}
	return nil
}

// Metrics returns the client metrics.
//
func (c *Client) Metrics() (opensearchtransport.Metrics, error) {
	if mt, ok := c.Transport.(opensearchtransport.Measurable); ok {
		return mt.Metrics()
	}
	return opensearchtransport.Metrics{}, errors.New("transport is missing method Metrics()")
}

// DiscoverNodes reloads the client connections by fetching information from the cluster.
//
func (c *Client) DiscoverNodes() error {
	if dt, ok := c.Transport.(opensearchtransport.Discoverable); ok {
		return dt.DiscoverNodes()
	}
	return errors.New("transport is missing method DiscoverNodes()")
}

// addrsFromEnvironment returns a list of addresses by splitting
// the given environment variable with comma, or an empty list.
//
func addrsFromEnvironment(name string) []string {
	var addrs []string

	if envURLs, ok := os.LookupEnv(name); ok && envURLs != "" {
		list := strings.Split(envURLs, ",")
		for _, u := range list {
			addrs = append(addrs, strings.TrimSpace(u))
		}
	}

	return addrs
}

// addrsToURLs creates a list of url.URL structures from url list.
//
func addrsToURLs(addrs []string) ([]*url.URL, error) {
	var urls []*url.URL
	for _, addr := range addrs {
		u, err := url.Parse(strings.TrimRight(addr, "/"))
		if err != nil {
			return nil, fmt.Errorf("cannot parse url: %v", err)
		}

		urls = append(urls, u)
	}
	return urls, nil
}

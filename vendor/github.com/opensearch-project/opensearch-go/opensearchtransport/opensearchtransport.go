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
	"bytes"
	"compress/gzip"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/opensearch-project/opensearch-go/signer"

	"github.com/opensearch-project/opensearch-go/internal/version"
)

const (
	// Version returns the package version as a string.
	Version = version.Client

	// esCompatHeader defines the env var for Compatibility header.
	esCompatHeader = "ELASTIC_CLIENT_APIVERSIONING"
)

var (
	userAgent           string
	compatibilityHeader bool
	reGoVersion         = regexp.MustCompile(`go(\d+\.\d+\..+)`)

	defaultMaxRetries    = 3
	defaultRetryOnStatus = [...]int{502, 503, 504}
)

func init() {
	userAgent = initUserAgent()

	compatHeaderEnv := os.Getenv(esCompatHeader)
	compatibilityHeader, _ = strconv.ParseBool(compatHeaderEnv)
}

// Interface defines the interface for HTTP client.
//
type Interface interface {
	Perform(*http.Request) (*http.Response, error)
}

// Config represents the configuration of HTTP client.
//
type Config struct {
	URLs     []*url.URL
	Username string
	Password string

	Header http.Header
	CACert []byte

	Signer signer.Signer

	RetryOnStatus        []int
	DisableRetry         bool
	EnableRetryOnTimeout bool
	MaxRetries           int
	RetryBackoff         func(attempt int) time.Duration

	CompressRequestBody bool

	EnableMetrics     bool
	EnableDebugLogger bool

	DiscoverNodesInterval time.Duration

	Transport http.RoundTripper
	Logger    Logger
	Selector  Selector

	ConnectionPoolFunc func([]*Connection, Selector) ConnectionPool
}

// Client represents the HTTP client.
//
type Client struct {
	sync.Mutex

	urls     []*url.URL
	username string
	password string
	header   http.Header

	signer signer.Signer

	retryOnStatus         []int
	disableRetry          bool
	enableRetryOnTimeout  bool
	maxRetries            int
	retryBackoff          func(attempt int) time.Duration
	discoverNodesInterval time.Duration
	discoverNodesTimer    *time.Timer

	compressRequestBody bool

	metrics *metrics

	transport http.RoundTripper
	logger    Logger
	selector  Selector
	pool      ConnectionPool
	poolFunc  func([]*Connection, Selector) ConnectionPool
}

// New creates new transport client.
//
// http.DefaultTransport will be used if no transport is passed in the configuration.
//
func New(cfg Config) (*Client, error) {
	if cfg.Transport == nil {
		cfg.Transport = http.DefaultTransport
	}

	if cfg.CACert != nil {
		httpTransport, ok := cfg.Transport.(*http.Transport)
		if !ok {
			return nil, fmt.Errorf("unable to set CA certificate for transport of type %T", cfg.Transport)
		}

		httpTransport = httpTransport.Clone()
		httpTransport.TLSClientConfig.RootCAs = x509.NewCertPool()

		if ok := httpTransport.TLSClientConfig.RootCAs.AppendCertsFromPEM(cfg.CACert); !ok {
			return nil, errors.New("unable to add CA certificate")
		}

		cfg.Transport = httpTransport
	}

	if len(cfg.RetryOnStatus) == 0 {
		cfg.RetryOnStatus = defaultRetryOnStatus[:]
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = defaultMaxRetries
	}

	var conns []*Connection
	for _, u := range cfg.URLs {
		conns = append(conns, &Connection{URL: u})
	}

	client := Client{
		urls:     cfg.URLs,
		username: cfg.Username,
		password: cfg.Password,
		header:   cfg.Header,

		signer: cfg.Signer,

		retryOnStatus:         cfg.RetryOnStatus,
		disableRetry:          cfg.DisableRetry,
		enableRetryOnTimeout:  cfg.EnableRetryOnTimeout,
		maxRetries:            cfg.MaxRetries,
		retryBackoff:          cfg.RetryBackoff,
		discoverNodesInterval: cfg.DiscoverNodesInterval,

		compressRequestBody: cfg.CompressRequestBody,

		transport: cfg.Transport,
		logger:    cfg.Logger,
		selector:  cfg.Selector,
		poolFunc:  cfg.ConnectionPoolFunc,
	}

	if client.poolFunc != nil {
		client.pool = client.poolFunc(conns, client.selector)
	} else {
		client.pool, _ = NewConnectionPool(conns, client.selector)
	}

	if cfg.EnableDebugLogger {
		debugLogger = &debuggingLogger{Output: os.Stdout}
	}

	if cfg.EnableMetrics {
		client.metrics = &metrics{responses: make(map[int]int)}
		// TODO(karmi): Type assertion to interface
		if pool, ok := client.pool.(*singleConnectionPool); ok {
			pool.metrics = client.metrics
		}
		if pool, ok := client.pool.(*statusConnectionPool); ok {
			pool.metrics = client.metrics
		}
	}

	if client.discoverNodesInterval > 0 {
		time.AfterFunc(client.discoverNodesInterval, func() {
			client.scheduleDiscoverNodes(client.discoverNodesInterval)
		})
	}

	return &client, nil
}

// Perform executes the request and returns a response or error.
//
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	var (
		res *http.Response
		err error
	)

	// Compatibility Header
	if compatibilityHeader {
		if req.Body != nil {
			req.Header.Set("Content-Type", "application/vnd.elasticsearch+json;compatible-with=7")
		}
		req.Header.Set("Accept", "application/vnd.elasticsearch+json;compatible-with=7")
	}

	// Record metrics, when enabled
	if c.metrics != nil {
		c.metrics.Lock()
		c.metrics.requests++
		c.metrics.Unlock()
	}

	// Update request
	c.setReqUserAgent(req)
	c.setReqGlobalHeader(req)

	if req.Body != nil && req.Body != http.NoBody {
		if c.compressRequestBody {
			var buf bytes.Buffer
			zw := gzip.NewWriter(&buf)
			if _, err := io.Copy(zw, req.Body); err != nil {
				return nil, fmt.Errorf("failed to compress request body: %s", err)
			}
			if err := zw.Close(); err != nil {
				return nil, fmt.Errorf("failed to compress request body (during close): %s", err)
			}

			req.GetBody = func() (io.ReadCloser, error) {
				r := buf
				return ioutil.NopCloser(&r), nil
			}
			req.Body, _ = req.GetBody()

			req.Header.Set("Content-Encoding", "gzip")
			req.ContentLength = int64(buf.Len())

		} else if req.GetBody == nil {
			if !c.disableRetry || (c.logger != nil && c.logger.RequestBodyEnabled()) {
				var buf bytes.Buffer
				buf.ReadFrom(req.Body)

				req.GetBody = func() (io.ReadCloser, error) {
					r := buf
					return ioutil.NopCloser(&r), nil
				}
				req.Body, _ = req.GetBody()
			}
		}
	}

	for i := 0; i <= c.maxRetries; i++ {
		var (
			conn            *Connection
			shouldRetry     bool
			shouldCloseBody bool
		)

		// Get connection from the pool
		c.Lock()
		conn, err = c.pool.Next()
		c.Unlock()
		if err != nil {
			if c.logger != nil {
				c.logRoundTrip(req, nil, err, time.Time{}, time.Duration(0))
			}
			return nil, fmt.Errorf("cannot get connection: %s", err)
		}

		// Update request
		c.setReqURL(conn.URL, req)
		c.setReqAuth(conn.URL, req)

		if err = c.signRequest(req); err != nil {
			return nil, fmt.Errorf("failed to sign request: %s", err)
		}

		if !c.disableRetry && i > 0 && req.Body != nil && req.Body != http.NoBody {
			body, err := req.GetBody()
			if err != nil {
				return nil, fmt.Errorf("cannot get request body: %s", err)
			}
			req.Body = body
		}

		// Set up time measures and execute the request
		start := time.Now().UTC()
		res, err = c.transport.RoundTrip(req)
		dur := time.Since(start)

		// Log request and response
		if c.logger != nil {
			if c.logger.RequestBodyEnabled() && req.Body != nil && req.Body != http.NoBody {
				req.Body, _ = req.GetBody()
			}
			c.logRoundTrip(req, res, err, start, dur)
		}

		if err != nil {
			// Record metrics, when enabled
			if c.metrics != nil {
				c.metrics.Lock()
				c.metrics.failures++
				c.metrics.Unlock()
			}

			// Report the connection as unsuccessful
			c.Lock()
			c.pool.OnFailure(conn)
			c.Unlock()

			// Retry on EOF errors
			if err == io.EOF {
				shouldRetry = true
			}

			// Retry on network errors, but not on timeout errors, unless configured
			if err, ok := err.(net.Error); ok {
				if (!err.Timeout() || c.enableRetryOnTimeout) && !c.disableRetry {
					shouldRetry = true
				}
			}
		} else {
			// Report the connection as succesfull
			c.Lock()
			c.pool.OnSuccess(conn)
			c.Unlock()
		}

		if res != nil && c.metrics != nil {
			c.metrics.Lock()
			c.metrics.responses[res.StatusCode]++
			c.metrics.Unlock()
		}

		// Retry on configured response statuses
		if res != nil && !c.disableRetry {
			for _, code := range c.retryOnStatus {
				if res.StatusCode == code {
					shouldRetry = true
					shouldCloseBody = true
				}
			}
		}

		// Break if retry should not be performed
		if !shouldRetry {
			break
		}

		// Drain and close body when retrying after response
		if shouldCloseBody && i < c.maxRetries {
			if res.Body != nil {
				io.Copy(ioutil.Discard, res.Body)
				res.Body.Close()
			}
		}

		// Delay the retry if a backoff function is configured
		if c.retryBackoff != nil {
			time.Sleep(c.retryBackoff(i + 1))
		}
	}

	// TODO(karmi): Wrap error
	return res, err
}

// URLs returns a list of transport URLs.
//
//
func (c *Client) URLs() []*url.URL {
	return c.pool.URLs()
}

func (c *Client) setReqURL(u *url.URL, req *http.Request) *http.Request {
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host

	if u.Path != "" {
		var b strings.Builder
		b.Grow(len(u.Path) + len(req.URL.Path))
		b.WriteString(u.Path)
		b.WriteString(req.URL.Path)
		req.URL.Path = b.String()
	}

	return req
}

func (c *Client) setReqAuth(u *url.URL, req *http.Request) *http.Request {
	if _, ok := req.Header["Authorization"]; !ok {
		if u.User != nil {
			password, _ := u.User.Password()
			req.SetBasicAuth(u.User.Username(), password)
			return req
		}

		if c.username != "" && c.password != "" {
			req.SetBasicAuth(c.username, c.password)
			return req
		}
	}

	return req
}

func (c *Client) signRequest(req *http.Request) error {
	if c.signer != nil {
		return c.signer.SignRequest(req)
	}
	return nil
}

func (c *Client) setReqUserAgent(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", userAgent)
	return req
}

func (c *Client) setReqGlobalHeader(req *http.Request) *http.Request {
	if len(c.header) > 0 {
		for k, v := range c.header {
			if req.Header.Get(k) != k {
				for _, vv := range v {
					req.Header.Add(k, vv)
				}
			}
		}
	}
	return req
}

func (c *Client) logRoundTrip(
	req *http.Request,
	res *http.Response,
	err error,
	start time.Time,
	dur time.Duration,
) {
	var dupRes http.Response
	if res != nil {
		dupRes = *res
	}
	if c.logger.ResponseBodyEnabled() {
		if res != nil && res.Body != nil && res.Body != http.NoBody {
			b1, b2, _ := duplicateBody(res.Body)
			dupRes.Body = b1
			res.Body = b2
		}
	}
	c.logger.LogRoundTrip(req, &dupRes, err, start, dur) // errcheck exclude
}

func initUserAgent() string {
	var b strings.Builder

	b.WriteString("opensearch-go")
	b.WriteRune('/')
	b.WriteString(Version)
	b.WriteRune(' ')
	b.WriteRune('(')
	b.WriteString(runtime.GOOS)
	b.WriteRune(' ')
	b.WriteString(runtime.GOARCH)
	b.WriteString("; ")
	b.WriteString("Go ")
	if v := reGoVersion.ReplaceAllString(runtime.Version(), "$1"); v != "" {
		b.WriteString(v)
	} else {
		b.WriteString(runtime.Version())
	}
	b.WriteRune(')')

	return b.String()
}

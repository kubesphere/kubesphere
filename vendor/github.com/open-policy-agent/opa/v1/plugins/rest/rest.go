// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package rest implements a REST client for communicating with remote services.
package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httputil"
	"reflect"
	"strings"

	"github.com/open-policy-agent/opa/internal/version"
	"github.com/open-policy-agent/opa/v1/keys"
	"github.com/open-policy-agent/opa/v1/logging"
	"github.com/open-policy-agent/opa/v1/tracing"
	"github.com/open-policy-agent/opa/v1/util"
)

const (
	defaultResponseHeaderTimeoutSeconds = int64(10)
	defaultResponseSizeLimitBytes       = 1024

	grantTypeClientCredentials = "client_credentials"
	grantTypeJwtBearer         = "jwt_bearer"
)

var maskedHeaderKeys = map[string]struct{}{
	"Authorization":        {},
	"X-Amz-Security-Token": {},
}

// An HTTPAuthPlugin represents a mechanism to construct and configure HTTP authentication for a REST service
type HTTPAuthPlugin interface {
	// implementations can assume NewClient will be called before Prepare
	NewClient(Config) (*http.Client, error)
	Prepare(*http.Request) error
}

// Config represents configuration for a REST client.
type Config struct {
	Name                         string            `json:"name"`
	URL                          string            `json:"url"`
	Headers                      map[string]string `json:"headers"`
	AllowInsecureTLS             bool              `json:"allow_insecure_tls,omitempty"`
	ResponseHeaderTimeoutSeconds *int64            `json:"response_header_timeout_seconds,omitempty"`
	TLS                          *serverTLSConfig  `json:"tls,omitempty"`
	Credentials                  struct {
		Bearer               *bearerAuthPlugin                  `json:"bearer,omitempty"`
		OAuth2               *oauth2ClientCredentialsAuthPlugin `json:"oauth2,omitempty"`
		ClientTLS            *clientTLSAuthPlugin               `json:"client_tls,omitempty"`
		S3Signing            *awsSigningAuthPlugin              `json:"s3_signing,omitempty"`
		GCPMetadata          *gcpMetadataAuthPlugin             `json:"gcp_metadata,omitempty"`
		AzureManagedIdentity *azureManagedIdentitiesAuthPlugin  `json:"azure_managed_identity,omitempty"`
		Plugin               *string                            `json:"plugin,omitempty"`
	} `json:"credentials"`
	Type   string `json:"type,omitempty"`
	keys   map[string]*keys.Config
	logger logging.Logger
}

// Equal returns true if this client config is equal to the other.
func (c *Config) Equal(other *Config) bool {
	otherWithoutLogger := *other
	otherWithoutLogger.logger = c.logger
	return reflect.DeepEqual(c, &otherWithoutLogger)
}

// An AuthPluginLookupFunc can lookup auth plugins by their name.
type AuthPluginLookupFunc func(name string) HTTPAuthPlugin

// AuthPlugin should be used to get an authentication method from the config.
func (c *Config) AuthPlugin(lookup AuthPluginLookupFunc) (HTTPAuthPlugin, error) {
	var candidate HTTPAuthPlugin
	if c.Credentials.Plugin != nil {
		if lookup == nil {
			// if no authPluginLookup function is passed we can't resolve the plugin
			return nil, errors.New("missing auth plugin lookup function")
		}

		candidate := lookup(*c.Credentials.Plugin)
		if candidate == nil {
			return nil, fmt.Errorf("auth plugin %q not found", *c.Credentials.Plugin)
		}

		return candidate, nil
	}
	// reflection avoids need for this code to change as auth plugins are added
	s := reflect.ValueOf(c.Credentials)
	for i := range s.NumField() {
		if s.Field(i).IsNil() {
			continue
		}

		if candidate != nil {
			return nil, errors.New("a maximum one credential method must be specified")
		}

		candidate = s.Field(i).Interface().(HTTPAuthPlugin)
	}

	if candidate == nil {
		return &defaultAuthPlugin{}, nil
	}
	return candidate, nil
}

func (c *Config) authHTTPClient(lookup AuthPluginLookupFunc) (*http.Client, error) {
	plugin, err := c.AuthPlugin(lookup)
	if err != nil {
		return nil, err
	}
	return plugin.NewClient(*c)
}

func (c *Config) authPrepare(req *http.Request, lookup AuthPluginLookupFunc) error {
	plugin, err := c.AuthPlugin(lookup)
	if err != nil {
		return err
	}
	return plugin.Prepare(req)
}

// Client implements an HTTP/REST client for communicating with remote
// services.
type Client struct {
	bytes                 *[]byte
	json                  *interface{}
	config                Config
	headers               map[string]string
	authPluginLookup      AuthPluginLookupFunc
	logger                logging.Logger
	loggerFields          map[string]interface{}
	distributedTacingOpts tracing.Options
}

// Name returns an option that overrides the service name on the client.
func Name(s string) func(*Client) {
	return func(c *Client) {
		c.config.Name = s
	}
}

// AuthPluginLookup assigns a function to lookup an HTTPAuthPlugin to a new Client.
// It's intended to be used when creating a Client using New(). Usually this is passed
// the plugins.AuthPlugin func, which retrieves a registered HTTPAuthPlugin from the
// plugin manager.
func AuthPluginLookup(l AuthPluginLookupFunc) func(*Client) {
	return func(c *Client) {
		c.authPluginLookup = l
	}
}

// Logger assigns a logger to the client
func Logger(l logging.Logger) func(*Client) {
	return func(c *Client) {
		c.logger = l
	}
}

// DistributedTracingOpts sets the options to be used by distributed tracing.
func DistributedTracingOpts(tr tracing.Options) func(*Client) {
	return func(c *Client) {
		c.distributedTacingOpts = tr
	}
}

// New returns a new Client for config.
func New(config []byte, keys map[string]*keys.Config, opts ...func(*Client)) (Client, error) {
	var parsedConfig Config
	if err := util.Unmarshal(config, &parsedConfig); err != nil {
		return Client{}, err
	}

	parsedConfig.URL = strings.TrimRight(parsedConfig.URL, "/")

	if parsedConfig.ResponseHeaderTimeoutSeconds == nil {
		timeout := defaultResponseHeaderTimeoutSeconds
		parsedConfig.ResponseHeaderTimeoutSeconds = &timeout
	}

	parsedConfig.keys = keys

	client := Client{
		config: parsedConfig,
	}

	for _, f := range opts {
		f(&client)
	}

	if client.logger == nil {
		client.logger = logging.Get()
	}

	client.config.logger = client.logger

	return client, nil
}

// AuthPluginLookup returns the lookup function to find a custom registered
// auth plugin by its name.
func (c Client) AuthPluginLookup() AuthPluginLookupFunc {
	return c.authPluginLookup
}

// Service returns the name of the service this Client is configured for.
func (c Client) Service() string {
	return c.config.Name
}

// Config returns this Client's configuration
func (c Client) Config() *Config {
	return &c.config
}

// SetResponseHeaderTimeout sets the "ResponseHeaderTimeout" in the http client's Transport
func (c Client) SetResponseHeaderTimeout(timeout *int64) Client {
	c.config.ResponseHeaderTimeoutSeconds = timeout
	return c
}

// Logger returns the logger assigned to the Client
func (c Client) Logger() logging.Logger {
	return c.logger
}

// LoggerFields returns the fields used for log statements used by Client
func (c Client) LoggerFields() map[string]interface{} {
	return c.loggerFields
}

// WithHeader returns a shallow copy of the client with a header to include the
// requests.
func (c Client) WithHeader(k, v string) Client {
	if v == "" {
		return c
	}
	if c.headers == nil {
		c.headers = map[string]string{}
	}
	c.headers[k] = v
	return c
}

// WithJSON returns a shallow copy of the client with the JSON value set as the
// message body to include the requests. This function sets the Content-Type
// header.
func (c Client) WithJSON(body interface{}) Client {
	c = c.WithHeader("Content-Type", "application/json")
	c.json = &body
	return c
}

// WithBytes returns a shallow copy of the client with the bytes set as the
// message body to include in the requests.
func (c Client) WithBytes(body []byte) Client {
	c.bytes = &body
	return c
}

// Do executes a request using the client.
func (c Client) Do(ctx context.Context, method, path string) (*http.Response, error) {

	httpClient, err := c.config.authHTTPClient(c.authPluginLookup)
	if err != nil {
		return nil, err
	}

	if len(c.distributedTacingOpts) > 0 {
		httpClient.Transport = tracing.NewTransport(httpClient.Transport, c.distributedTacingOpts)
	}

	path = strings.Trim(path, "/")

	var body io.Reader

	if c.bytes != nil {
		body = bytes.NewReader(*c.bytes)
	} else if c.json != nil {
		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(*c.json); err != nil {
			return nil, err
		}
		body = &buf
	}

	url := c.config.URL + "/" + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"User-Agent": version.UserAgent,
	}

	// Copy custom headers from config.
	maps.Copy(headers, c.config.Headers)

	// Overwrite with headers set directly on client.
	maps.Copy(headers, c.headers)

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if err = c.config.authPrepare(req, c.authPluginLookup); err != nil {
		return nil, err
	}

	if c.logger.GetLevel() >= logging.Debug {
		c.loggerFields = map[string]interface{}{
			"method":  method,
			"url":     url,
			"headers": withMaskedHeaders(req.Header),
		}

		c.logger.WithFields(c.loggerFields).Debug("Sending request.")
	}

	resp, err := httpClient.Do(req)

	if resp != nil && c.logger.GetLevel() >= logging.Debug {
		// Only log for debug purposes. If an error occurred, the caller should handle
		// that. In the non-error case, the caller may not do anything.
		c.loggerFields["status"] = resp.Status
		c.loggerFields["headers"] = resp.Header

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			dump, err := httputil.DumpResponse(resp, true)
			if err != nil {
				return nil, err
			}

			if len(dump) < defaultResponseSizeLimitBytes {
				c.loggerFields["response"] = string(dump)
			} else {
				c.loggerFields["response"] = fmt.Sprintf("%v...", string(dump[:defaultResponseSizeLimitBytes]))
			}
		}
		c.logger.WithFields(c.loggerFields).Debug("Received response.")
	}

	return resp, err
}

func withMaskedHeaders(headers http.Header) http.Header {
	masked := make(http.Header)
	for k, v := range headers {
		if _, ok := maskedHeaderKeys[k]; ok {
			masked.Set(k, "REDACTED")
		} else {
			masked[k] = v
		}
	}
	return masked
}

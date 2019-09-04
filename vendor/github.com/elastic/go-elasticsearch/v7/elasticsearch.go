package elasticsearch

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/elastic/go-elasticsearch/v7/esapi"
	"github.com/elastic/go-elasticsearch/v7/estransport"
	"github.com/elastic/go-elasticsearch/v7/internal/version"
)

const (
	defaultURL = "http://localhost:9200"
)

// Version returns the package version as a string.
//
const Version = version.Client

// Config represents the client configuration.
//
type Config struct {
	Addresses []string // A list of Elasticsearch nodes to use.
	Username  string   // Username for HTTP Basic Authentication.
	Password  string   // Password for HTTP Basic Authentication.

	CloudID string // Endpoint for the Elastic Service (https://elastic.co/cloud).
	APIKey  string // Base64-encoded token for authorization; if set, overrides username and password.

	Transport http.RoundTripper  // The HTTP transport object.
	Logger    estransport.Logger // The logger object.
}

// Client represents the Elasticsearch client.
//
type Client struct {
	*esapi.API // Embeds the API methods
	Transport  estransport.Interface
}

// NewDefaultClient creates a new client with default options.
//
// It will use http://localhost:9200 as the default address.
//
// It will use the ELASTICSEARCH_URL environment variable, if set,
// to configure the addresses; use a comma to separate multiple URLs.
//
func NewDefaultClient() (*Client, error) {
	return NewClient(Config{})
}

// NewClient creates a new client with configuration from cfg.
//
// It will use http://localhost:9200 as the default address.
//
// It will use the ELASTICSEARCH_URL environment variable, if set,
// to configure the addresses; use a comma to separate multiple URLs.
//
// It's an error to set both cfg.Addresses and the ELASTICSEARCH_URL
// environment variable.
//
func NewClient(cfg Config) (*Client, error) {
	var addrs []string

	envAddrs := addrsFromEnvironment()

	if len(envAddrs) > 0 && len(cfg.Addresses) > 0 {
		return nil, errors.New("cannot create client: both ELASTICSEARCH_URL and Addresses are set")
	}

	if len(envAddrs) > 0 && cfg.CloudID != "" {
		return nil, errors.New("cannot create client: both ELASTICSEARCH_URL and CloudID are set")
	}

	if len(cfg.Addresses) > 0 && cfg.CloudID != "" {
		return nil, errors.New("cannot create client: both Adresses and CloudID are set")
	}

	if cfg.CloudID != "" {
		cloudAddrs, err := addrFromCloudID(cfg.CloudID)
		if err != nil {
			return nil, fmt.Errorf("cannot create client: cannot parse CloudID: %s", err)
		}
		addrs = append(addrs, cloudAddrs)
	} else {
		if len(envAddrs) > 0 {
			addrs = append(envAddrs, envAddrs...)
		}
		if len(cfg.Addresses) > 0 {
			addrs = append(envAddrs, cfg.Addresses...)
		}
	}

	urls, err := addrsToURLs(addrs)
	if err != nil {
		return nil, fmt.Errorf("cannot create client: %s", err)
	}

	if len(urls) == 0 {
		u, _ := url.Parse(defaultURL) // errcheck exclude
		urls = append(urls, u)
	}

	tp := estransport.New(estransport.Config{
		URLs:     urls,
		Username: cfg.Username,
		Password: cfg.Password,
		APIKey:   cfg.APIKey,

		Transport: cfg.Transport,
		Logger:    cfg.Logger,
	})

	return &Client{Transport: tp, API: esapi.New(tp)}, nil
}

// Perform delegates to Transport to execute a request and return a response.
//
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	return c.Transport.Perform(req)
}

// addrsFromEnvironment returns a list of addresses by splitting
// the ELASTICSEARCH_URL environment variable with comma, or an empty list.
//
func addrsFromEnvironment() []string {
	var addrs []string

	if envURLs, ok := os.LookupEnv("ELASTICSEARCH_URL"); ok && envURLs != "" {
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

// addrFromCloudID extracts the Elasticsearch URL from CloudID.
// See: https://www.elastic.co/guide/en/cloud/current/ec-cloud-id.html
//
func addrFromCloudID(input string) (string, error) {
	var (
		port   = 9243
		scheme = "https://"
	)

	values := strings.Split(input, ":")
	if len(values) != 2 {
		return "", fmt.Errorf("unexpected format: %q", input)
	}
	data, err := base64.StdEncoding.DecodeString(values[1])
	if err != nil {
		return "", err
	}
	parts := strings.Split(string(data), "$")
	return fmt.Sprintf("%s%s.%s:%d", scheme, parts[1], parts[0], port), nil
}

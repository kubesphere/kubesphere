package estransport

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v7/internal/version"
)

// Version returns the package version as a string.
//
const Version = version.Client

var (
	userAgent   string
	reGoVersion = regexp.MustCompile(`go(\d+\.\d+\..+)`)
)

func init() {
	userAgent = initUserAgent()
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
	APIKey   string

	Transport http.RoundTripper
	Logger    Logger
}

// Client represents the HTTP client.
//
type Client struct {
	urls     []*url.URL
	username string
	password string
	apikey   string

	transport http.RoundTripper
	selector  Selector
	logger    Logger
}

// New creates new HTTP client.
//
// http.DefaultTransport will be used if no transport is passed in the configuration.
//
func New(cfg Config) *Client {
	if cfg.Transport == nil {
		cfg.Transport = http.DefaultTransport
	}

	return &Client{
		urls:     cfg.URLs,
		username: cfg.Username,
		password: cfg.Password,
		apikey:   cfg.APIKey,

		transport: cfg.Transport,
		selector:  NewRoundRobinSelector(cfg.URLs...),
		logger:    cfg.Logger,
	}
}

// Perform executes the request and returns a response or error.
//
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	var (
		dupReqBody io.Reader
	)

	// Get URL from the Selector
	//
	u, err := c.getURL()
	if err != nil {
		// TODO(karmi): Log error
		return nil, fmt.Errorf("cannot get URL: %s", err)
	}

	// Update request
	//
	c.setURL(u, req)
	c.setUserAgent(req)
	c.setAuthorization(u, req)

	// Duplicate request body for logger
	//
	if c.logger != nil && c.logger.RequestBodyEnabled() {
		if req.Body != nil && req.Body != http.NoBody {
			dupReqBody, req.Body, _ = duplicateBody(req.Body)
		}
	}

	// Set up time measures and execute the request
	//
	start := time.Now().UTC()
	res, err := c.transport.RoundTrip(req)
	dur := time.Since(start)

	// Log request and response
	//
	if c.logger != nil {
		c.logRoundTrip(req, res, dupReqBody, err, start, dur)
	}

	// TODO(karmi): Wrap error
	return res, err
}

// URLs returns a list of transport URLs.
//
func (c *Client) URLs() []*url.URL {
	return c.urls
}

func (c *Client) getURL() (*url.URL, error) {
	return c.selector.Select()
}

func (c *Client) setURL(u *url.URL, req *http.Request) *http.Request {
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

func (c *Client) setAuthorization(u *url.URL, req *http.Request) *http.Request {
	if _, ok := req.Header["Authorization"]; !ok {
		if u.User != nil {
			password, _ := u.User.Password()
			req.SetBasicAuth(u.User.Username(), password)
			return req
		}

		if c.apikey != "" {
			var b bytes.Buffer
			b.Grow(len("APIKey ") + len(c.apikey))
			b.WriteString("APIKey ")
			b.WriteString(c.apikey)
			req.Header.Set("Authorization", b.String())
			return req
		}

		if c.username != "" && c.password != "" {
			req.SetBasicAuth(c.username, c.password)
			return req
		}
	}

	return req
}

func (c *Client) setUserAgent(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", userAgent)
	return req
}

func (c *Client) logRoundTrip(
	req *http.Request,
	res *http.Response,
	reqBody io.Reader,
	err error,
	start time.Time,
	dur time.Duration,
) {
	var dupRes http.Response
	if res != nil {
		dupRes = *res
	}
	if c.logger.RequestBodyEnabled() {
		if req.Body != nil && req.Body != http.NoBody {
			req.Body = ioutil.NopCloser(reqBody)
		}
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

	b.WriteString("go-elasticsearch")
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

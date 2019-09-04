package estransport

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v6/internal/version"
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

	Transport http.RoundTripper
	Logger    Logger
}

// Client represents the HTTP client.
//
type Client struct {
	urls     []*url.URL
	username string
	password string

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

		transport: cfg.Transport,
		selector:  NewRoundRobinSelector(cfg.URLs...),
		logger:    cfg.Logger,
	}
}

// Perform executes the request and returns a response or error.
//
func (c *Client) Perform(req *http.Request) (*http.Response, error) {
	u, err := c.getURL()
	if err != nil {
		// TODO(karmi): Log error
		return nil, fmt.Errorf("cannot get URL: %s", err)
	}

	c.setURL(u, req)
	c.setUserAgent(req)

	if _, ok := req.Header["Authorization"]; !ok {
		c.setBasicAuth(u, req)
	}

	var dupReqBody *bytes.Buffer
	if c.logger != nil && c.logger.RequestBodyEnabled() {
		if req.Body != nil && req.Body != http.NoBody {
			dupReqBody = bytes.NewBuffer(make([]byte, 0, int(req.ContentLength)))
			dupReqBody.ReadFrom(req.Body)
			req.Body = ioutil.NopCloser(bytes.NewBuffer(dupReqBody.Bytes()))
		}
	}

	start := time.Now().UTC()
	res, err := c.transport.RoundTrip(req)
	dur := time.Since(start)

	if c.logger != nil {
		var dupRes http.Response
		if res != nil {
			dupRes = *res
		}
		if c.logger.RequestBodyEnabled() {
			if req.Body != nil && req.Body != http.NoBody {
				req.Body = ioutil.NopCloser(dupReqBody)
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

func (c *Client) setBasicAuth(u *url.URL, req *http.Request) *http.Request {
	if u.User != nil {
		password, _ := u.User.Password()
		req.SetBasicAuth(u.User.Username(), password)
		return req
	}

	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
		return req
	}

	return req
}

func (c *Client) setUserAgent(req *http.Request) *http.Request {
	req.Header.Set("User-Agent", userAgent)
	return req
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

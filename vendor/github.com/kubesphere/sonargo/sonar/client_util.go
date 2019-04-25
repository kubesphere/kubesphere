package sonargo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-querystring/query"
)

// SetBaseURL sets the base URL for API requests to a custom endpoint. urlStr
// should always be specified with a trailing slash.
func SetBaseURLUtil(urlStr string) (*url.URL, error) {
	// Make sure the given URL end with a slash
	if !strings.HasSuffix(urlStr, "/") {
		urlStr += "/"
	}

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	// Update the base URL of the client.
	return baseURL, nil
}

// NewRequest creates an API request. A relative URL path can be provided in
// urlStr, in which case it is resolved relative to the base URL of the Client.
// Relative URL paths should always be specified without a preceding slash. If
// specified, the value pointed to by body is JSON encoded and included as the
// request body.
func NewRequest(method, path string, baseURL *url.URL, username, password string, opt interface{}) (*http.Request, error) {
	// Set the encoded opaque data
	u := *baseURL
	unescaped, err := url.PathUnescape(path)
	if err != nil {
		return nil, err
	}
	u.RawPath = u.Path + path
	u.Path = u.Path + unescaped
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}

	req := &http.Request{
		Method: method,
		URL:    &u,
		Proto:  "HTTP/1.1",
		Header: make(http.Header),
		Host:   u.Host,
	}

	if method == "POST" || method == "PUT" {
		//SonarQube use RawQuery even method is POST
		bodyBytes, err := json.Marshal(opt)
		if err != nil {
			return nil, err
		}
		bodyReader := bytes.NewReader(bodyBytes)

		u.RawQuery = ""
		req.Body = ioutil.NopCloser(bodyReader)
		req.ContentLength = int64(bodyReader.Len())
		req.Header.Set("Content-Type", "application/json")
	}

	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(username, password)
	return req, nil
}

// Do sends an API request and returns the API response. The API response is
// JSON decoded and stored in the value pointed to by v, or returned as an
// error if an API error has occurred. If v implements the io.Writer
// interface, the raw response body will be written to v, without attempting to
// first decode it.
func Do(c *http.Client, req *http.Request, v interface{}) (*http.Response, error) {
	isText := false
	if _, ok := v.(*string); ok {
		req.Header.Set("Accept", "text/plain")
		isText = true
	}
	glog.V(1).Infof("[%s] %s\n", req.Method, req.URL.String())
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}
	if v != nil {
		defer resp.Body.Close()
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			if isText {
				byts, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return resp, err
				}
				w := v.(*string)
				*w = string(byts)
			} else {
				decoder := json.NewDecoder(resp.Body)
				err = decoder.Decode(v)
			}
		}
	}
	return resp, err
}

type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

func (e *ErrorResponse) Error() string {
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, e.Response.Request.URL.RequestURI())
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Body = data

		var raw interface{}
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = string(data)
		} else {
			errorResponse.Message = parseError(raw)
		}
	}

	return errorResponse
}
func parseError(raw interface{}) string {
	switch raw := raw.(type) {
	case string:
		return raw

	case []interface{}:
		var errs []string
		for _, v := range raw {
			errs = append(errs, parseError(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(errs, ", "))

	case map[string]interface{}:
		var errs []string
		for k, v := range raw {
			errs = append(errs, fmt.Sprintf("{%s: %s}", k, parseError(v)))
		}
		sort.Strings(errs)
		return strings.Join(errs, ", ")

	default:
		return fmt.Sprintf("failed to parse unexpected error type: %T", raw)
	}
}

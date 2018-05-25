package client // import "github.com/docker/docker/client"

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/gotestyourself/gotestyourself/assert"
)

// TestSetHostHeader should set fake host for local communications, set real host
// for normal communications.
func TestSetHostHeader(t *testing.T) {
	testURL := "/test"
	testCases := []struct {
		host            string
		expectedHost    string
		expectedURLHost string
	}{
		{
			"unix:///var/run/docker.sock",
			"docker",
			"/var/run/docker.sock",
		},
		{
			"npipe:////./pipe/docker_engine",
			"docker",
			"//./pipe/docker_engine",
		},
		{
			"tcp://0.0.0.0:4243",
			"",
			"0.0.0.0:4243",
		},
		{
			"tcp://localhost:4243",
			"",
			"localhost:4243",
		},
	}

	for c, test := range testCases {
		hostURL, err := ParseHostURL(test.host)
		assert.NilError(t, err)

		client := &Client{
			client: newMockClient(func(req *http.Request) (*http.Response, error) {
				if !strings.HasPrefix(req.URL.Path, testURL) {
					return nil, fmt.Errorf("Test Case #%d: Expected URL %q, got %q", c, testURL, req.URL)
				}
				if req.Host != test.expectedHost {
					return nil, fmt.Errorf("Test Case #%d: Expected host %q, got %q", c, test.expectedHost, req.Host)
				}
				if req.URL.Host != test.expectedURLHost {
					return nil, fmt.Errorf("Test Case #%d: Expected URL host %q, got %q", c, test.expectedURLHost, req.URL.Host)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewReader(([]byte("")))),
				}, nil
			}),

			proto:    hostURL.Scheme,
			addr:     hostURL.Host,
			basePath: hostURL.Path,
		}

		_, err = client.sendRequest(context.Background(), "GET", testURL, nil, nil, nil)
		assert.NilError(t, err)
	}
}

// TestPlainTextError tests the server returning an error in plain text for
// backwards compatibility with API versions <1.24. All other tests use
// errors returned as JSON
func TestPlainTextError(t *testing.T) {
	client := &Client{
		client: newMockClient(plainTextErrorMock(http.StatusInternalServerError, "Server error")),
	}
	_, err := client.ContainerList(context.Background(), types.ContainerListOptions{})
	if err == nil || err.Error() != "Error response from daemon: Server error" {
		t.Fatalf("expected a Server Error, got %v", err)
	}
}

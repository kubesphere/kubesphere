package kiali

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

type MockClient struct {
	TokenResult   []byte
	RequestResult string
}

func (c *MockClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(c.RequestResult))),
	}, nil
}

func (c *MockClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(c.TokenResult)),
	}, nil
}

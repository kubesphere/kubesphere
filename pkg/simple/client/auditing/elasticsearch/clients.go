/*
Copyright 2020 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package elasticsearch

import (
	"fmt"
	es5 "github.com/elastic/go-elasticsearch/v5"
	es5api "github.com/elastic/go-elasticsearch/v5/esapi"
	es6 "github.com/elastic/go-elasticsearch/v6"
	es6api "github.com/elastic/go-elasticsearch/v6/esapi"
	es7 "github.com/elastic/go-elasticsearch/v7"
	es7api "github.com/elastic/go-elasticsearch/v7/esapi"
	jsoniter "github.com/json-iterator/go"
	"io"
	"net/http"
)

type Request struct {
	Index string
	Body  io.Reader
}

type Response struct {
	Hits         Hits                           `json:"hits"`
	Aggregations map[string]jsoniter.RawMessage `json:"aggregations"`
}

type Hits struct {
	Total int64               `json:"total"`
	Hits  jsoniter.RawMessage `json:"hits"`
}

type Error struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
	Status int    `json:"status"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s %s: %s", http.StatusText(e.Status), e.Type, e.Reason)
}

type ClientV5 es5.Client

func (c *ClientV5) ExSearch(r *Request) (*Response, error) {
	return c.parse(c.Search(c.Search.WithIndex(r.Index), c.Search.WithBody(r.Body), c.Search.WithIgnoreUnavailable(true)))
}
func (c *ClientV5) parse(resp *es5api.Response, err error) (*Response, error) {
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	var r struct {
		Hits struct {
			Total int64               `json:"total"`
			Hits  jsoniter.RawMessage `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]jsoniter.RawMessage `json:"aggregations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}
	return &Response{
		Hits:         Hits{Total: r.Hits.Total, Hits: r.Hits.Hits},
		Aggregations: r.Aggregations,
	}, nil
}
func (c *ClientV5) Version() (string, error) {
	resp, err := c.Info()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.IsError() {
		return "", fmt.Errorf(resp.String())
	}
	var r map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("error parsing the response body: %s", err)
	}
	return fmt.Sprintf("%s", r["version"].(map[string]interface{})["number"]), nil
}

type ClientV6 es6.Client

func (c *ClientV6) ExSearch(r *Request) (*Response, error) {
	return c.parse(c.Search(c.Search.WithIndex(r.Index), c.Search.WithBody(r.Body), c.Search.WithIgnoreUnavailable(true)))
}
func (c *ClientV6) parse(resp *es6api.Response, err error) (*Response, error) {
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	var r struct {
		Hits *struct {
			Total int64               `json:"total"`
			Hits  jsoniter.RawMessage `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]jsoniter.RawMessage `json:"aggregations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}
	return &Response{
		Hits:         Hits{Total: r.Hits.Total, Hits: r.Hits.Hits},
		Aggregations: r.Aggregations,
	}, nil
}

type ClientV7 es7.Client

func (c *ClientV7) ExSearch(r *Request) (*Response, error) {
	return c.parse(c.Search(c.Search.WithIndex(r.Index), c.Search.WithBody(r.Body), c.Search.WithIgnoreUnavailable(true)))
}
func (c *ClientV7) parse(resp *es7api.Response, err error) (*Response, error) {
	if err != nil {
		return nil, fmt.Errorf("error getting response: %s", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.IsError() {
		return nil, fmt.Errorf(resp.String())
	}
	var r struct {
		Hits *struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits jsoniter.RawMessage `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]jsoniter.RawMessage `json:"aggregations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("error parsing the response body: %s", err)
	}
	return &Response{
		Hits:         Hits{Total: r.Hits.Total.Value, Hits: r.Hits.Hits},
		Aggregations: r.Aggregations,
	}, nil
}

type client interface {
	ExSearch(r *Request) (*Response, error)
}

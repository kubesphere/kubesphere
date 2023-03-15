/*
Copyright 2021 KubeSphere Authors

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

package kiali

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"kubesphere.io/kubesphere/pkg/simple/client/cache"
)

// Kiali token Response
type TokenResponse struct {
	// The username for the token
	Username string `json:"username"`
	// The authentication token
	Token string `json:"token"`
	// The expired time for the token
	ExpiresOn string `json:"expiresOn"`
}

// Kiali Authentication  Strategy
type Strategy string

const (
	AuthStrategyToken     Strategy = "token"
	AuthStrategyAnonymous Strategy = "anonymous"
)

const (
	AuthURL            = "%s/kiali/api/authenticate"
	KialiTokenCacheKey = "kubesphere:kubesphere:kiali"
)

type HttpClient interface {
	// Do is an interface of http client Do method,
	// that sends an HTTP request and returns an HTTP response.
	Do(req *http.Request) (*http.Response, error)

	// PostForm is an interface of http client PostForm method,
	// that issues a POST to the specified URL.
	PostForm(url string, data url.Values) (resp *http.Response, err error)
}

// Kiali Client
type Client struct {
	Strategy     Strategy
	cache        cache.Interface
	client       HttpClient
	ServiceToken string
	Host         string
}

// NewClient creates an instance of Kiali Client.
func NewClient(strategy Strategy,
	cache cache.Interface,
	client HttpClient,
	serviceToken string,
	host string) *Client {

	return &Client{
		Strategy:     strategy,
		cache:        cache,
		client:       client,
		ServiceToken: serviceToken,
		Host:         host,
	}

}

// NewDefaultClient creates an instance of Kiali Client with default http settings.
func NewDefaultClient(
	cache cache.Interface,
	serviceToken string,
	host string) *Client {
	return &Client{
		Strategy:     AuthStrategyToken,
		cache:        cache,
		client:       &http.Client{},
		ServiceToken: serviceToken,
		Host:         host,
	}
}

// authenticate sends auth request with Kubernetes token and
// get Kiali token from the response.
func (c *Client) authenticate() (*TokenResponse, error) {
	resp, err := c.client.PostForm(fmt.Sprintf(AuthURL, c.Host), url.Values{
		"token": {c.ServiceToken},
	})
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	token := TokenResponse{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		return nil, err
	}
	// token strategy in kiali:v1.46 writes the token in the cookie
	// Related issue: https://github.com/kiali/kiali/issues/4682
	token.Token = resp.Header.Get("Set-Cookie")

	return &token, nil
}

// Get issues a GET to the Kiali server with the url.
func (c *Client) Get(url string) (resp *http.Response, err error) {

	if req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", c.Host, url), nil); err != nil {
		return nil, err
	} else {
		if c.Strategy == AuthStrategyToken {
			err := c.SetToken(req)
			if err != nil {
				return nil, err
			}
		}
		resp, err := c.client.Do(req)

		if err != nil {
			c.clearTokenCache(err)
		}

		return resp, err
	}
}

func (c *Client) clearTokenCache(err error) {
	if c.cache != nil && err != nil {
		c.cache.Del(KialiTokenCacheKey)
	}
}

// SetToken gets token from the Kiali server/cache and sets Bearer token to the request header.
func (c *Client) SetToken(req *http.Request) error {
	if c.cache != nil {
		token, err := c.cache.Get(KialiTokenCacheKey)
		if err == nil {
			req.Header.Set("Cookie", token)
			return nil
		}
	}

	token, err := c.authenticate()
	if err != nil {
		return err
	}
	// token strategy in kiali:v1.46 writes the token in the cookie.
	// https://github.com/kiali/kiali-operator/blob/v1.50.1/molecule/asserts/token-test/assert-token-access.yml#L47-L56
	req.Header.Set("Cookie", token.Token)

	if c.cache != nil {
		c.cache.Set(KialiTokenCacheKey, token.Token, time.Hour)
	}
	return nil
}

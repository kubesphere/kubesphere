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

package cas

import (
	"crypto/tls"
	"fmt"
	"github.com/mitchellh/mapstructure"
	gocas "gopkg.in/cas.v2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"net/http"
	"net/url"
)

func init() {
	identityprovider.RegisterOAuthProvider(&casProviderFactory{})
}

type cas struct {
	RedirectURL        string `json:"redirectURL" yaml:"redirectURL"`
	CASServerURL       string `json:"casServerURL" yaml:"casServerURL"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	client             *gocas.RestClient
}

type casProviderFactory struct {
}

type casIdentity struct {
	User string `json:"user"`
}

func (c casIdentity) GetUserID() string {
	return c.User
}

func (c casIdentity) GetUsername() string {
	return c.User
}

func (c casIdentity) GetEmail() string {
	return ""
}

func (f casProviderFactory) Type() string {
	return "CASIdentityProvider"
}

func (f casProviderFactory) Create(options oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var cas cas
	if err := mapstructure.Decode(options, &cas); err != nil {
		return nil, err
	}
	casURL, err := url.Parse(cas.CASServerURL)
	if err != nil {
		return nil, err
	}
	redirectURL, err := url.Parse(cas.RedirectURL)
	if err != nil {
		return nil, err
	}
	cas.client = gocas.NewRestClient(&gocas.RestOptions{
		CasURL:     casURL,
		ServiceURL: redirectURL,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cas.InsecureSkipVerify},
			},
		},
		URLScheme: nil,
	})
	return &cas, nil
}

func (c cas) IdentityExchange(ticket string) (identityprovider.Identity, error) {
	resp, err := c.client.ValidateServiceTicket(gocas.ServiceTicket(ticket))
	if err != nil {
		return nil, fmt.Errorf("cas validate service ticket failed: %v", err)
	}
	return &casIdentity{User: resp.User}, nil
}

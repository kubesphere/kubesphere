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

package aliyunidaas

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

type AliyunIDaaS struct {
	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"-" yaml:"clientSecret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint Endpoint `json:"endpoint" yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirectURL" yaml:"redirectURL"`

	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`
}

// Endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type Endpoint struct {
	AuthURL     string `json:"authURL" yaml:"authURL"`
	TokenURL    string `json:"tokenURL" yaml:"tokenURL"`
	UserInfoURL string `json:"user_info_url" yaml:"userInfoUrl"`
}

type IDaaSIdentity struct {
	Sub         string `json:"sub"`
	OuID        string `json:"ou_id"`
	Nickname    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	OuName      string `json:"ou_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type UserInfoResp struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	Code          string        `json:"code"`
	IDaaSIdentity IDaaSIdentity `json:"data"`
}

func init() {
	identityprovider.RegisterOAuthProvider(&AliyunIDaaS{})
}

func (a *AliyunIDaaS) Type() string {
	return "AliyunIDaasProvider"
}

func (a *AliyunIDaaS) Setup(options *oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	data, err := yaml.Marshal(options)
	if err != nil {
		return nil, err
	}
	var provider AliyunIDaaS
	err = yaml.Unmarshal(data, &provider)
	if err != nil {
		return nil, err
	}
	return &provider, nil
}

func (a IDaaSIdentity) GetName() string {
	return a.Username
}

func (a IDaaSIdentity) GetEmail() string {
	return a.Email
}

func (g *AliyunIDaaS) IdentityExchange(code string) (identityprovider.Identity, error) {
	config := oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   g.Endpoint.AuthURL,
			TokenURL:  g.Endpoint.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		RedirectURL: g.RedirectURL,
		Scopes:      g.Scopes,
	}
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	resp, err := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token)).Get(g.Endpoint.UserInfoURL)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)

	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	var UserInfoResp UserInfoResp
	err = json.Unmarshal(data, &UserInfoResp)
	if err != nil {
		return nil, err
	}

	if !UserInfoResp.Success {
		return nil, errors.New(UserInfoResp.Message)
	}

	return UserInfoResp.IDaaSIdentity, nil
}

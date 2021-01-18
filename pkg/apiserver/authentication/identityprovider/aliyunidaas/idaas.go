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
	"github.com/mitchellh/mapstructure"
	"io/ioutil"

	"golang.org/x/oauth2"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
)

func init() {
	identityprovider.RegisterOAuthProvider(&idaasProviderFactory{})
}

type aliyunIDaaS struct {
	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"-" yaml:"clientSecret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.Endpoint.
	Endpoint endpoint `json:"endpoint" yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirectURL" yaml:"redirectURL"`

	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`
}

// endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type endpoint struct {
	AuthURL     string `json:"authURL" yaml:"authURL"`
	TokenURL    string `json:"tokenURL" yaml:"tokenURL"`
	UserInfoURL string `json:"user_info_url" yaml:"userInfoUrl"`
}

type idaasIdentity struct {
	Sub         string `json:"sub"`
	OuID        string `json:"ou_id"`
	Nickname    string `json:"nickname"`
	PhoneNumber string `json:"phone_number"`
	OuName      string `json:"ou_name"`
	Email       string `json:"email"`
	Username    string `json:"username"`
}

type userInfoResp struct {
	Success       bool          `json:"success"`
	Message       string        `json:"message"`
	Code          string        `json:"code"`
	IDaaSIdentity idaasIdentity `json:"data"`
}

type idaasProviderFactory struct {
}

func (g *idaasProviderFactory) Type() string {
	return "AliyunIDaasProvider"
}

func (g *idaasProviderFactory) Create(options *oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var idaas aliyunIDaaS
	if err := mapstructure.Decode(options, &idaas); err != nil {
		return nil, err
	}
	return &idaas, nil
}

func (a idaasIdentity) GetUserID() string {
	return a.Sub
}

func (a idaasIdentity) GetUsername() string {
	return a.Username
}

func (a idaasIdentity) GetEmail() string {
	return a.Email
}

func (a idaasIdentity) GetDisplayName() string {
	return a.Nickname
}

func (a *aliyunIDaaS) IdentityExchange(code string) (identityprovider.Identity, error) {
	config := oauth2.Config{
		ClientID:     a.ClientID,
		ClientSecret: a.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   a.Endpoint.AuthURL,
			TokenURL:  a.Endpoint.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		RedirectURL: a.RedirectURL,
		Scopes:      a.Scopes,
	}
	token, err := config.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}

	resp, err := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token)).Get(a.Endpoint.UserInfoURL)
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var UserInfoResp userInfoResp
	err = json.Unmarshal(data, &UserInfoResp)
	if err != nil {
		return nil, err
	}

	if !UserInfoResp.Success {
		return nil, errors.New(UserInfoResp.Message)
	}

	return UserInfoResp.IDaaSIdentity, nil
}

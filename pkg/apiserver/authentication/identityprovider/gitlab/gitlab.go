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

package gitlab

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

const (
	userInfoURL = "https://gitlab.com/api/v4/user"
	authURL     = "https://gitlab.com/oauth/authorize"
	tokenURL    = "https://gitlab.com/oauth/token"
)

func init() {
	identityprovider.RegisterOAuthProvider(&gitlabProviderFactory{})
}

type gitlab struct {
	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"-" yaml:"clientSecret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or sso.endpoint.
	Endpoint endpoint `json:"endpoint" yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirectURL" yaml:"redirectURL"`

	// Used to turn off TLS certificate checks
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`

	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`

	Config *oauth2.Config `json:"-" yaml:"-"`
}

// endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type endpoint struct {
	AuthURL     string `json:"authURL" yaml:"authURL"`
	TokenURL    string `json:"tokenURL" yaml:"tokenURL"`
	UserInfoURL string `json:"userInfoURL" yaml:"userInfoURL"`
}

type gitlabIdentity struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	State     string `json:"state"`
	AvatarURL string `json:"avatar_url"`
	WebURL    string `json:"web_url"`
}

type gitlabProviderFactory struct {
}

func (g *gitlabProviderFactory) Type() string {
	return "GitlabIdentityProvider"
}

func (g *gitlabProviderFactory) Create(opts options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var gitlab gitlab
	if err := mapstructure.Decode(opts, &gitlab); err != nil {
		return nil, err
	}

	if gitlab.Endpoint.AuthURL == "" {
		gitlab.Endpoint.AuthURL = authURL
	}
	if gitlab.Endpoint.TokenURL == "" {
		gitlab.Endpoint.TokenURL = tokenURL
	}
	if gitlab.Endpoint.UserInfoURL == "" {
		gitlab.Endpoint.UserInfoURL = userInfoURL
	}
	// fixed options
	opts["endpoint"] = options.DynamicOptions{
		"authURL":     gitlab.Endpoint.AuthURL,
		"tokenURL":    gitlab.Endpoint.TokenURL,
		"userInfoURL": gitlab.Endpoint.UserInfoURL,
	}
	gitlab.Config = &oauth2.Config{
		ClientID:     gitlab.ClientID,
		ClientSecret: gitlab.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  gitlab.Endpoint.AuthURL,
			TokenURL: gitlab.Endpoint.TokenURL,
		},
		RedirectURL: gitlab.RedirectURL,
		Scopes:      gitlab.Scopes,
	}
	return &gitlab, nil
}

func (g gitlabIdentity) GetUserID() string {
	return strconv.FormatInt(g.ID, 10)
}

func (g gitlabIdentity) GetUsername() string {
	return g.Username
}

func (g gitlabIdentity) GetEmail() string {
	return g.Email
}

func (g *gitlab) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	// OAuth2 callback, see also https://tools.ietf.org/html/rfc6749#section-4.1.2
	code := req.URL.Query().Get("code")
	ctx := req.Context()
	if g.InsecureSkipVerify {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	}
	token, err := g.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	resp, err := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)).Get(g.Endpoint.UserInfoURL)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var gitlabIdentity gitlabIdentity
	err = json.Unmarshal(data, &gitlabIdentity)
	if err != nil {
		return nil, err
	}

	return gitlabIdentity, nil
}

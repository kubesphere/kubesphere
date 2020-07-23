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

package oauth

import (
	"errors"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"net/url"
	"time"
)

type GrantHandlerType string
type MappingMethod string
type IdentityProviderType string

const (
	// GrantHandlerAuto auto-approves client authorization grant requests
	GrantHandlerAuto GrantHandlerType = "auto"
	// GrantHandlerPrompt prompts the user to approve new client authorization grant requests
	GrantHandlerPrompt GrantHandlerType = "prompt"
	// GrantHandlerDeny auto-denies client authorization grant requests
	GrantHandlerDeny GrantHandlerType = "deny"
	// MappingMethodAuto  The default value.The user will automatically create and mapping when login successful.
	// Fails if a user with that user name is already mapped to another identity.
	MappingMethodAuto MappingMethod = "auto"
	// MappingMethodLookup Looks up an existing identity, user identity mapping, and user, but does not automatically
	// provision users or identities. Using this method requires you to manually provision users.
	MappingMethodLookup MappingMethod = "lookup"
	// MappingMethodMixed  A user entity can be mapped with multiple identifyProvider.
	MappingMethodMixed MappingMethod = "mixed"
)

var (
	ErrorClientNotFound        = errors.New("the OAuth client was not found")
	ErrorProviderNotFound      = errors.New("the identity provider was not found")
	ErrorRedirectURLNotAllowed = errors.New("redirect URL is not allowed")
)

type Options struct {
	// Register identity providers.
	IdentityProviders []IdentityProviderOptions `json:"identityProviders,omitempty" yaml:"identityProviders,omitempty"`

	// Register additional OAuth clients.
	Clients []Client `json:"clients,omitempty" yaml:"clients,omitempty"`

	// AccessTokenMaxAgeSeconds  control the lifetime of access tokens. The default lifetime is 24 hours.
	// 0 means no expiration.
	AccessTokenMaxAge time.Duration `json:"accessTokenMaxAge" yaml:"accessTokenMaxAge"`

	// Inactivity timeout for tokens
	// The value represents the maximum amount of time that can occur between
	// consecutive uses of the token. Tokens become invalid if they are not
	// used within this temporal window. The user will need to acquire a new
	// token to regain access once a token times out.
	// This value needs to be set only if the default set in configuration is
	// not appropriate for this client. Valid values are:
	// - 0: Tokens for this client never time out
	// - X: Tokens time out if there is no activity
	// The current minimum allowed value for X is 5 minutes
	AccessTokenInactivityTimeout time.Duration `json:"accessTokenInactivityTimeout" yaml:"accessTokenInactivityTimeout"`
}

type DynamicOptions map[string]interface{}

type IdentityProviderOptions struct {
	// The provider name.
	Name string `json:"name" yaml:"name"`

	// Defines how new identities are mapped to users when they login. Allowed values are:
	//  - auto:   The default value.The user will automatically create and mapping when login successful.
	//            Fails if a user with that user name is already mapped to another identity.
	//  - lookup: Looks up an existing identity, user identity mapping, and user, but does not automatically
	//            provision users or identities. Using this method requires you to manually provision users.
	//  - mixed:  A user entity can be mapped with multiple identifyProvider.
	MappingMethod MappingMethod `json:"mappingMethod" yaml:"mappingMethod"`

	// The type of identify provider
	// OpenIDIdentityProvider LDAPIdentityProvider GitHubIdentityProvider
	Type string `json:"type" yaml:"type"`

	// The options of identify provider
	Provider *DynamicOptions `json:"-" yaml:"provider"`
}

type Token struct {
	// AccessToken is the token that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// TokenType is the type of token.
	// The Type method returns either this or "Bearer", the default.
	TokenType string `json:"token_type,omitempty"`

	// RefreshToken is a token that's used by the application
	// (as opposed to the user) to refresh the access token
	// if it expires.
	RefreshToken string `json:"refresh_token,omitempty"`

	// ExpiresIn is the optional expiration second of the access token.
	ExpiresIn int `json:"expires_in,omitempty"`
}

type Client struct {
	// The name of the OAuth client is used as the client_id parameter when making requests to <master>/oauth/authorize
	// and <master>/oauth/token.
	Name string

	// Secret is the unique secret associated with a client
	Secret string `json:"-" yaml:"secret,omitempty"`

	// RespondWithChallenges indicates whether the client wants authentication needed responses made
	// in the form of challenges instead of redirects
	RespondWithChallenges bool `json:"respondWithChallenges,omitempty" yaml:"respondWithChallenges,omitempty"`

	// RedirectURIs is the valid redirection URIs associated with a client
	RedirectURIs []string `json:"redirectURIs,omitempty" yaml:"redirectURIs,omitempty"`

	// GrantMethod determines how to handle grants for this client. If no method is provided, the
	// cluster default grant handling method will be used. Valid grant handling methods are:
	//  - auto:   always approves grant requests, useful for trusted clients
	//  - prompt: prompts the end user for approval of grant requests, useful for third-party clients
	//  - deny:   always denies grant requests, useful for black-listed clients
	GrantMethod GrantHandlerType `json:"grantMethod,omitempty" yaml:"grantMethod,omitempty"`

	// ScopeRestrictions describes which scopes this client can request.  Each requested scope
	// is checked against each restriction.  If any restriction matches, then the scope is allowed.
	// If no restriction matches, then the scope is denied.
	ScopeRestrictions []string `json:"scopeRestrictions,omitempty" yaml:"scopeRestrictions,omitempty"`

	// AccessTokenMaxAge overrides the default access token max age for tokens granted to this client.
	AccessTokenMaxAge *time.Duration `json:"accessTokenMaxAge,omitempty" yaml:"accessTokenMaxAge,omitempty"`

	// AccessTokenInactivityTimeout overrides the default token
	// inactivity timeout for tokens granted to this client.
	AccessTokenInactivityTimeout *time.Duration `json:"accessTokenInactivityTimeout,omitempty" yaml:"accessTokenInactivityTimeout,omitempty"`
}

var (
	// Allow any redirect URI if the redirectURI is defined in request
	AllowAllRedirectURI                 = "*"
	DefaultTokenMaxAge                  = time.Second * 86400
	DefaultAccessTokenInactivityTimeout = time.Duration(0)
	DefaultClients                      = []Client{{
		Name:                         "default",
		Secret:                       "kubesphere",
		RespondWithChallenges:        true,
		RedirectURIs:                 []string{AllowAllRedirectURI},
		GrantMethod:                  GrantHandlerAuto,
		ScopeRestrictions:            []string{"full"},
		AccessTokenMaxAge:            &DefaultTokenMaxAge,
		AccessTokenInactivityTimeout: &DefaultAccessTokenInactivityTimeout,
	}}
)

func (o *Options) OAuthClient(name string) (Client, error) {
	for _, found := range o.Clients {
		if found.Name == name {
			return found, nil
		}
	}
	for _, defaultClient := range DefaultClients {
		if defaultClient.Name == name {
			return defaultClient, nil
		}
	}
	return Client{}, ErrorClientNotFound
}
func (o *Options) IdentityProviderOptions(name string) (*IdentityProviderOptions, error) {
	for _, found := range o.IdentityProviders {
		if found.Name == name {
			return &found, nil
		}
	}
	return nil, ErrorProviderNotFound
}

func (c Client) anyRedirectAbleURI() []string {
	uris := make([]string, 0)
	for _, uri := range c.RedirectURIs {
		_, err := url.Parse(uri)
		if err == nil {
			uris = append(uris, uri)
		}
	}
	return uris
}

func (c Client) ResolveRedirectURL(expectURL string) (string, error) {
	// RedirectURIs is empty
	if len(c.RedirectURIs) == 0 {
		return "", ErrorRedirectURLNotAllowed
	}
	allowAllRedirectURI := sliceutil.HasString(c.RedirectURIs, AllowAllRedirectURI)
	redirectAbleURIs := c.anyRedirectAbleURI()

	if expectURL == "" {
		// Need to specify at least one RedirectURI
		if len(redirectAbleURIs) > 0 {
			return redirectAbleURIs[0], nil
		} else {
			return "", ErrorRedirectURLNotAllowed
		}
	}
	if allowAllRedirectURI || sliceutil.HasString(redirectAbleURIs, expectURL) {
		return expectURL, nil
	}

	return "", ErrorRedirectURLNotAllowed
}

func NewOptions() *Options {
	return &Options{
		IdentityProviders:            make([]IdentityProviderOptions, 0),
		Clients:                      make([]Client, 0),
		AccessTokenMaxAge:            time.Hour * 2,
		AccessTokenInactivityTimeout: time.Hour * 2,
	}
}

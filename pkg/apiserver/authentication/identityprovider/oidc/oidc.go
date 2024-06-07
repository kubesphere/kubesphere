/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oidc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

func init() {
	identityprovider.RegisterOAuthProviderFactory(&oidcProviderFactory{})
}

type oidcProvider struct {
	// Defines how Clients dynamically discover information about OpenID Providers
	// See also, https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfig
	Issuer string `json:"issuer,omitempty" yaml:"issuer,omitempty"`

	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"-" yaml:"clientSecret"`

	// Endpoint contains the resource server's token endpoint URLs.
	// These are constants specific to each server and are often available via site-specific packages,
	// such as google.Endpoint or github.Endpoint.
	Endpoint endpoint `json:"endpoint" yaml:"endpoint"`

	// RedirectURL is the URL to redirect users going through
	// the OAuth flow, after the resource owner's URLs.
	RedirectURL string `json:"redirectURL" yaml:"redirectURL"`

	// Scope specifies optional requested permissions.
	Scopes []string `json:"scopes" yaml:"scopes"`

	// GetUserInfo uses the userinfo endpoint to get additional claims for the token.
	// This is especially useful where upstreams return "thin" id tokens
	// See also, https://openid.net/specs/openid-connect-core-1_0.html#UserInfo
	GetUserInfo bool `json:"getUserInfo" yaml:"getUserInfo"`

	// Used to turn off TLS certificate checks
	InsecureSkipVerify bool `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`

	// Configurable key which contains the email claims
	EmailKey string `json:"emailKey" yaml:"emailKey"`

	// Configurable key which contains the preferred username claims
	PreferredUsernameKey string `json:"preferredUsernameKey" yaml:"preferredUsernameKey"`

	Provider     *oidc.Provider        `json:"-" yaml:"-"`
	OAuth2Config *oauth2.Config        `json:"-" yaml:"-"`
	Verifier     *oidc.IDTokenVerifier `json:"-" yaml:"-"`
}

// endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type endpoint struct {
	// URL of the OP's OAuth 2.0 Authorization Endpoint [OpenID.Core](https://openid.net/specs/openid-connect-discovery-1_0.html#OpenID.Core).
	AuthURL string `json:"authURL" yaml:"authURL"`
	// URL of the OP's OAuth 2.0 Token Endpoint [OpenID.Core](https://openid.net/specs/openid-connect-discovery-1_0.html#OpenID.Core).
	// This is REQUIRED unless only the Implicit Flow is used.
	TokenURL string `json:"tokenURL" yaml:"tokenURL"`
	// URL of the OP's UserInfo Endpoint [OpenID.Core](https://openid.net/specs/openid-connect-discovery-1_0.html#OpenID.Core).
	// This URL MUST use the https scheme and MAY contain port, path, and query parameter components.
	UserInfoURL string `json:"userInfoURL" yaml:"userInfoURL"`
	//  URL of the OP's JSON Web Key Set [JWK](https://openid.net/specs/openid-connect-discovery-1_0.html#JWK) document.
	JWKSURL string `json:"jwksURL"`
	// URL at the OP to which an RP can perform a redirect to request that the End-User be logged out at the OP.
	// This URL MUST use the https scheme and MAY contain port, path, and query parameter components.
	// https://openid.net/specs/openid-connect-rpinitiated-1_0.html#OPMetadata
	EndSessionURL string `json:"endSessionURL"`
}

type oidcIdentity struct {
	// Subject - Identifier for the End-User at the Issuer.
	Sub string `json:"sub"`
	// Shorthand name by which the End-User wishes to be referred to at the RP,
	// such as janedoe or j.doe. This value MAY be any valid JSON string including special characters such as @, /, or whitespace.
	// The RP MUST NOT rely upon this value being unique
	PreferredUsername string `json:"preferred_username"`
	// End-User's preferred e-mail address.
	// Its value MUST conform to the RFC 5322 [RFC5322] addr-spec syntax.
	// The RP MUST NOT rely upon this value being unique.
	Email string `json:"email"`
}

func (o oidcIdentity) GetUserID() string {
	return o.Sub
}

func (o oidcIdentity) GetUsername() string {
	return o.PreferredUsername
}

func (o oidcIdentity) GetEmail() string {
	return o.Email
}

type oidcProviderFactory struct {
}

func (f *oidcProviderFactory) Type() string {
	return "OIDCIdentityProvider"
}

func (f *oidcProviderFactory) Create(opts options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var oidcProvider oidcProvider
	if err := mapstructure.Decode(opts, &oidcProvider); err != nil {
		return nil, err
	}
	// dynamically discover
	if oidcProvider.Issuer != "" {
		ctx := context.TODO()
		if oidcProvider.InsecureSkipVerify {
			client := &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: true,
					},
				},
			}
			ctx = oidc.ClientContext(ctx, client)
		}
		provider, err := oidc.NewProvider(ctx, oidcProvider.Issuer)
		if err != nil {
			return nil, fmt.Errorf("failed to create oidc provider: %v", err)
		}
		var providerJSON map[string]interface{}
		if err = provider.Claims(&providerJSON); err != nil {
			return nil, fmt.Errorf("failed to decode oidc provider claims: %v", err)
		}
		oidcProvider.Endpoint.AuthURL, _ = providerJSON["authorization_endpoint"].(string)
		oidcProvider.Endpoint.TokenURL, _ = providerJSON["token_endpoint"].(string)
		oidcProvider.Endpoint.UserInfoURL, _ = providerJSON["userinfo_endpoint"].(string)
		oidcProvider.Endpoint.JWKSURL, _ = providerJSON["jwks_uri"].(string)
		oidcProvider.Endpoint.EndSessionURL, _ = providerJSON["end_session_endpoint"].(string)
		oidcProvider.Provider = provider
		oidcProvider.Verifier = provider.Verifier(&oidc.Config{
			// TODO: support HS256
			ClientID: oidcProvider.ClientID,
		})
		opts["endpoint"] = options.DynamicOptions{
			"authURL":       oidcProvider.Endpoint.AuthURL,
			"tokenURL":      oidcProvider.Endpoint.TokenURL,
			"userInfoURL":   oidcProvider.Endpoint.UserInfoURL,
			"jwksURL":       oidcProvider.Endpoint.JWKSURL,
			"endSessionURL": oidcProvider.Endpoint.EndSessionURL,
		}
	}
	scopes := []string{oidc.ScopeOpenID}
	if !sliceutil.HasString(oidcProvider.Scopes, oidc.ScopeOpenID) {
		scopes = append(scopes, oidcProvider.Scopes...)
	}
	oidcProvider.Scopes = scopes
	oidcProvider.OAuth2Config = &oauth2.Config{
		ClientID:     oidcProvider.ClientID,
		ClientSecret: oidcProvider.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: oidcProvider.Endpoint.TokenURL,
			AuthURL:  oidcProvider.Endpoint.AuthURL,
		},
		RedirectURL: oidcProvider.RedirectURL,
		Scopes:      oidcProvider.Scopes,
	}

	return &oidcProvider, nil
}

func (o *oidcProvider) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	//OAuth2 callback, see also https://tools.ietf.org/html/rfc6749#section-4.1.2
	code := req.URL.Query().Get("code")
	ctx := req.Context()
	if o.InsecureSkipVerify {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		ctx = context.WithValue(ctx, oauth2.HTTPClient, client)
	}
	token, err := o.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("oidc: failed to get token: %v", err)
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token in token response")
	}
	var claims jwt.MapClaims
	if o.Verifier != nil {
		idToken, err := o.Verifier.Verify(ctx, rawIDToken)
		if err != nil {
			return nil, fmt.Errorf("failed to verify id token: %v", err)
		}
		if err := idToken.Claims(&claims); err != nil {
			return nil, fmt.Errorf("failed to decode id token claims: %v", err)
		}
	} else {
		_, _, err := new(jwt.Parser).ParseUnverified(rawIDToken, &claims)
		if err != nil {
			return nil, fmt.Errorf("failed to decode id token claims: %v", err)
		}
		if err := claims.Valid(); err != nil {
			return nil, fmt.Errorf("failed to verify id token: %v", err)
		}
	}
	if o.GetUserInfo {
		if o.Provider != nil {
			userInfo, err := o.Provider.UserInfo(ctx, oauth2.StaticTokenSource(token))
			if err != nil {
				return nil, fmt.Errorf("failed to fetch userinfo: %v", err)
			}
			if err := userInfo.Claims(&claims); err != nil {
				return nil, fmt.Errorf("failed to decode userinfo claims: %v", err)
			}
		} else {
			resp, err := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)).Get(o.Endpoint.UserInfoURL)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch userinfo: %v", err)
			}
			data, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch userinfo: %v", err)
			}
			_ = resp.Body.Close()
			if err := json.Unmarshal(data, &claims); err != nil {
				return nil, fmt.Errorf("failed to decode userinfo claims: %v", err)
			}
		}
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return nil, errors.New("missing required claim \"sub\"")
	}

	var email string
	emailKey := "email"
	if o.EmailKey != "" {
		emailKey = o.EmailKey
	}
	email, _ = claims[emailKey].(string)

	var preferredUsername string
	preferredUsernameKey := "preferred_username"
	if o.PreferredUsernameKey != "" {
		preferredUsernameKey = o.PreferredUsernameKey
	}

	preferredUsername, _ = claims[preferredUsernameKey].(string)
	if preferredUsername == "" {
		preferredUsername, _ = claims["name"].(string)
	}

	return &oidcIdentity{
		Sub:               subject,
		PreferredUsername: preferredUsername,
		Email:             email,
	}, nil
}

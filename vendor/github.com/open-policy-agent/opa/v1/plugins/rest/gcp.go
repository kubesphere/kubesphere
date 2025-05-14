// Copyright 2020 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	defaultGCPMetadataEndpoint = "http://metadata.google.internal"
	defaultAccessTokenPath     = "/computeMetadata/v1/instance/service-accounts/default/token"
	defaultIdentityTokenPath   = "/computeMetadata/v1/instance/service-accounts/default/identity"
)

// AccessToken holds a GCP access token.
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type gcpMetadataError struct {
	err        error
	endpoint   string
	statusCode int
}

func (e *gcpMetadataError) Error() string {
	return fmt.Sprintf("error retrieving gcp ID token from %s %d: %v", e.endpoint, e.statusCode, e.err)
}

func (e *gcpMetadataError) Unwrap() error { return e.err }

var (
	errGCPMetadataNotFound       = errors.New("not found")
	errGCPMetadataInvalidRequest = errors.New("invalid request")
	errGCPMetadataUnexpected     = errors.New("unexpected error")
)

// gcpMetadataAuthPlugin represents authentication via GCP metadata service.
type gcpMetadataAuthPlugin struct {
	AccessTokenPath   string   `json:"access_token_path"`
	Audience          string   `json:"audience"`
	Endpoint          string   `json:"endpoint"`
	IdentityTokenPath string   `json:"identity_token_path"`
	Scopes            []string `json:"scopes"`
}

func (ap *gcpMetadataAuthPlugin) NewClient(c Config) (*http.Client, error) {
	if ap.Audience == "" && len(ap.Scopes) == 0 {
		return nil, errors.New("audience or scopes is required when gcp metadata is enabled")
	}

	if ap.Audience != "" && len(ap.Scopes) > 0 {
		return nil, errors.New("either audience or scopes can be set, not both, when gcp metadata is enabled")
	}

	if ap.Endpoint == "" {
		ap.Endpoint = defaultGCPMetadataEndpoint
	}

	if ap.AccessTokenPath == "" {
		ap.AccessTokenPath = defaultAccessTokenPath
	}

	if ap.IdentityTokenPath == "" {
		ap.IdentityTokenPath = defaultIdentityTokenPath
	}

	t, err := DefaultTLSConfig(c)
	if err != nil {
		return nil, err
	}

	return DefaultRoundTripperClient(t, *c.ResponseHeaderTimeoutSeconds), nil
}

func (ap *gcpMetadataAuthPlugin) Prepare(req *http.Request) error {
	var err error
	var token string

	if ap.Audience != "" {
		token, err = identityTokenFromMetadataService(ap.Endpoint, ap.IdentityTokenPath, ap.Audience)
		if err != nil {
			return fmt.Errorf("error retrieving identity token from gcp metadata service: %w", err)
		}
	}

	if len(ap.Scopes) != 0 {
		token, err = accessTokenFromMetadataService(ap.Endpoint, ap.AccessTokenPath, ap.Scopes)
		if err != nil {
			return fmt.Errorf("error retrieving access token from gcp metadata service: %w", err)
		}
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %v", token))
	return nil
}

// accessTokenFromMetadataService returns an access token based on the scopes.
func accessTokenFromMetadataService(endpoint, path string, scopes []string) (string, error) {
	s := strings.Join(scopes, ",")

	e := fmt.Sprintf("%s%s?scopes=%s", endpoint, path, s)

	data, err := gcpMetadataServiceRequest(e)
	if err != nil {
		return "", err
	}

	var accessToken AccessToken
	err = json.Unmarshal(data, &accessToken)
	if err != nil {
		return "", err
	}

	return accessToken.AccessToken, nil
}

// identityTokenFromMetadataService returns an identity token based on the audience.
func identityTokenFromMetadataService(endpoint, path, audience string) (string, error) {
	e := fmt.Sprintf("%s%s?audience=%s", endpoint, path, audience)

	data, err := gcpMetadataServiceRequest(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func gcpMetadataServiceRequest(endpoint string) ([]byte, error) {
	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Metadata-Flavor", "Google")

	timeout := time.Duration(5) * time.Second
	httpClient := http.Client{Timeout: timeout}

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	switch s := response.StatusCode; s {
	case 200:
		break
	case 400:
		return nil, &gcpMetadataError{errGCPMetadataInvalidRequest, endpoint, s}
	case 404:
		return nil, &gcpMetadataError{errGCPMetadataNotFound, endpoint, s}
	default:
		return nil, &gcpMetadataError{errGCPMetadataUnexpected, endpoint, s}
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

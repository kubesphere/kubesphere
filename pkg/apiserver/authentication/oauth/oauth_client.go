/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"gopkg.in/yaml.v3"

	v1 "k8s.io/api/core/v1"
	errorsutil "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

const (
	GrantMethodAuto   = "auto"
	GrantMethodPrompt = "prompt"
	GrantMethodDeny   = "deny"

	ConfigTypeOAuthClient = "oauthclient"
	SecretTypeOAuthClient = "config.kubesphere.io/" + ConfigTypeOAuthClient
	SecretDataKey         = "configuration.yaml"
)

var (
	ErrorClientNotFound        = errors.New("the OAuth client was not found")
	ErrorRedirectURLNotAllowed = errors.New("redirect URL is not allowed")

	ValidGrantMethods = []string{GrantMethodAuto, GrantMethodPrompt, GrantMethodDeny}
)

// Client represents an OAuth client configuration.
type Client struct {
	// Name is the unique identifier for the OAuth client. It is used as the client_id parameter
	// when making requests to <master>/oauth/authorize.
	Name string `json:"name" yaml:"name"`

	// Secret is the unique secret associated with the client for secure communication.
	Secret string `json:"-" yaml:"secret"`

	// Trusted indicates whether the client is considered a trusted client.
	Trusted bool `json:"trusted" yaml:"trusted"`

	// GrantMethod determines how grant requests for this client should be handled. If no method is provided,
	// the cluster default grant handling method will be used. Valid grant handling methods are:
	//   - auto: Always approves grant requests, useful for trusted clients.
	//   - prompt: Prompts the end user for approval of grant requests, useful for third-party clients.
	//   - deny: Always denies grant requests, useful for black-listed clients.
	GrantMethod string `json:"grantMethod" yaml:"grantMethod"`

	// RespondWithChallenges indicates whether the client prefers authentication needed responses
	// in the form of challenges instead of redirects.
	RespondWithChallenges bool `json:"respondWithChallenges,omitempty" yaml:"respondWithChallenges,omitempty"`

	// ScopeRestrictions describes which scopes this client can request. Each requested scope
	// is checked against each restriction. If any restriction matches, then the scope is allowed.
	// If no restriction matches, then the scope is denied.
	ScopeRestrictions []string `json:"scopeRestrictions,omitempty" yaml:"scopeRestrictions,omitempty"`

	// RedirectURIs is a list of valid redirection URIs associated with the client.
	RedirectURIs []string `json:"redirectURIs,omitempty" yaml:"redirectURIs,omitempty"`

	// AccessTokenMaxAge overrides the default maximum age for access tokens granted to this client.
	// The default value is 7200 seconds, and the minimum allowed value is 600 seconds.
	AccessTokenMaxAgeSeconds int64 `json:"accessTokenMaxAgeSeconds,omitempty" yaml:"accessTokenMaxAgeSeconds,omitempty"`

	// AccessTokenInactivityTimeout overrides the default token inactivity timeout
	// for tokens granted to this client.
	AccessTokenInactivityTimeoutSeconds int64 `json:"accessTokenInactivityTimeoutSeconds,omitempty" yaml:"accessTokenInactivityTimeoutSeconds,omitempty"`
}

type ClientGetter interface {
	GetOAuthClient(ctx context.Context, name string) (*Client, error)
	ListOAuthClients(ctx context.Context) ([]*Client, error)
}

func NewOAuthClientGetter(reader client.Reader) ClientGetter {
	return &oauthClientGetter{reader}
}

type oauthClientGetter struct {
	client.Reader
}

func (o *oauthClientGetter) ListOAuthClients(ctx context.Context) ([]*Client, error) {
	clients := make([]*Client, 0)
	secrets := &v1.SecretList{}
	if err := o.List(ctx, secrets, client.InNamespace(constants.KubeSphereNamespace),
		client.MatchingLabels{constants.GenericConfigTypeLabel: ConfigTypeOAuthClient}); err != nil {
		return nil, err
	}
	for _, secret := range secrets.Items {
		if secret.Type != SecretTypeOAuthClient {
			continue
		}
		if c, err := UnmarshalFrom(&secret); err != nil {
			klog.Errorf("failed to unmarshal secret data: %s", err)
			continue
		} else {
			clients = append(clients, c)
		}
	}
	return clients, nil
}

// GetOAuthClient retrieves an OAuth client by name from the underlying storage.
// It returns the OAuth client if found; otherwise, returns an error.
func (o *oauthClientGetter) GetOAuthClient(ctx context.Context, name string) (*Client, error) {
	clients, err := o.ListOAuthClients(ctx)
	if err != nil {
		klog.Errorf("failed to list OAuth clients: %v", err)
		return nil, err
	}
	for _, c := range clients {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, ErrorClientNotFound
}

// ValidateClient validates the properties of the provided OAuth 2.0 client.
// It checks the client's grant method, access token inactivity timeout, and access
// token max age for validity. If any validation fails, it returns an aggregated error.
func ValidateClient(client Client) error {
	var validationErrors []error

	// Validate grant method.
	if !sliceutil.HasString(ValidGrantMethods, client.GrantMethod) {
		validationErrors = append(validationErrors, fmt.Errorf("invalid grant method: %s", client.GrantMethod))
	}

	// Validate access token inactivity timeout.
	if client.AccessTokenInactivityTimeoutSeconds != 0 && client.AccessTokenInactivityTimeoutSeconds < 600 {
		validationErrors = append(validationErrors, fmt.Errorf("invalid access token inactivity timeout: %d, the minimum value can only be 600", client.AccessTokenInactivityTimeoutSeconds))
	}

	// Validate access token max age.
	if client.AccessTokenMaxAgeSeconds != 0 && client.AccessTokenMaxAgeSeconds < 600 {
		validationErrors = append(validationErrors, fmt.Errorf("invalid access token max age: %d, the minimum value can only be 600", client.AccessTokenMaxAgeSeconds))
	}

	// Aggregate validation errors and return.
	return errorsutil.NewAggregate(validationErrors)
}

// ResolveRedirectURL resolves the redirect URL for the OAuth 2.0 authorization process.
// It takes an expected URL as a parameter and returns the resolved URL if it's allowed.
// If the expected URL is not provided, it uses the first available RedirectURI from the client.
func (c *Client) ResolveRedirectURL(expectURL string) (*url.URL, error) {
	// Check if RedirectURIs are specified for the client.
	if len(c.RedirectURIs) == 0 {
		return nil, ErrorRedirectURLNotAllowed
	}

	// Get the list of redirectable URIs for the client.
	redirectAbleURIs := filterValidRedirectURIs(c.RedirectURIs)

	// If the expected URL is not provided, use the first available RedirectURI.
	if expectURL == "" {
		if len(redirectAbleURIs) > 0 {
			return url.Parse(redirectAbleURIs[0])
		} else {
			// No RedirectURIs available for the client.
			return nil, ErrorRedirectURLNotAllowed
		}
	}

	// Check if the provided expected URL is allowed.
	if sliceutil.HasString(redirectAbleURIs, expectURL) {
		return url.Parse(expectURL)
	}

	// The provided expected URL is not allowed.
	return nil, ErrorRedirectURLNotAllowed
}

// IsValidScope checks whether the requested scope is valid for the client.
// It compares each individual scope in the requested scope string with the client's
// allowed scope restrictions. If all scopes are allowed, it returns true; otherwise, false.
func (c *Client) IsValidScope(requestedScope string) bool {
	// Split the requested scope string into individual scopes.
	scopes := strings.Split(requestedScope, " ")

	// Check each individual scope against the client's scope restrictions.
	for _, scope := range scopes {
		if !sliceutil.HasString(c.ScopeRestrictions, scope) {
			// Log a message indicating the disallowed scope.
			klog.V(4).Infof("Invalid scope: %s is not allowed for client %s", scope, c.Name)
			return false
		}
	}

	// All scopes are valid.
	return true
}

// filterValidRedirectURIs filters out invalid redirect URIs from the given slice.
// It returns a new slice containing only valid URIs.
func filterValidRedirectURIs(redirectURIs []string) []string {
	validURIs := make([]string, 0)
	for _, uri := range redirectURIs {
		// Check if the URI is valid by attempting to parse it.
		_, err := url.Parse(uri)
		if err == nil {
			// The URI is valid, add it to the list of valid URIs.
			validURIs = append(validURIs, uri)
		}
	}
	return validURIs
}

func UnmarshalFrom(secret *v1.Secret) (*Client, error) {
	oc := &Client{}
	if err := yaml.Unmarshal(secret.Data[SecretDataKey], oc); err != nil {
		return nil, err
	}
	return oc, nil
}

func MarshalInto(client *Client, secret *v1.Secret) error {
	data, err := yaml.Marshal(client)
	if err != nil {
		return err
	}
	secret.Data = map[string][]byte{SecretDataKey: data}
	return nil
}

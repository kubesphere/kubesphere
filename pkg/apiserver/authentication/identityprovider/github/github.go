/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package github

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

const (
	userInfoURL = "https://api.github.com/user"
	authURL     = "https://github.com/login/oauth/authorize"
	tokenURL    = "https://github.com/login/oauth/access_token"
)

func init() {
	identityprovider.RegisterOAuthProviderFactory(&ldapProviderFactory{})
}

type github struct {
	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"-" yaml:"clientSecret"`

	// Endpoint contains the resource server's token endpoint
	// URLs. These are constants specific to each server and are
	// often available via site-specific packages, such as
	// google.Endpoint or github.endpoint.
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

type githubIdentity struct {
	Login             string    `json:"login"`
	ID                int       `json:"id"`
	NodeID            string    `json:"node_id"`
	AvatarURL         string    `json:"avatar_url"`
	GravatarID        string    `json:"gravatar_id"`
	URL               string    `json:"url"`
	HTMLURL           string    `json:"html_url"`
	FollowersURL      string    `json:"followers_url"`
	FollowingURL      string    `json:"following_url"`
	GistsURL          string    `json:"gists_url"`
	StarredURL        string    `json:"starred_url"`
	SubscriptionsURL  string    `json:"subscriptions_url"`
	OrganizationsURL  string    `json:"organizations_url"`
	ReposURL          string    `json:"repos_url"`
	EventsURL         string    `json:"events_url"`
	ReceivedEventsURL string    `json:"received_events_url"`
	Type              string    `json:"type"`
	SiteAdmin         bool      `json:"site_admin"`
	Name              string    `json:"name"`
	Company           string    `json:"company"`
	Blog              string    `json:"blog"`
	Location          string    `json:"location"`
	Email             string    `json:"email"`
	Hireable          bool      `json:"hireable"`
	Bio               string    `json:"bio"`
	PublicRepos       int       `json:"public_repos"`
	PublicGists       int       `json:"public_gists"`
	Followers         int       `json:"followers"`
	Following         int       `json:"following"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	PrivateGists      int       `json:"private_gists"`
	TotalPrivateRepos int       `json:"total_private_repos"`
	OwnedPrivateRepos int       `json:"owned_private_repos"`
	DiskUsage         int       `json:"disk_usage"`
	Collaborators     int       `json:"collaborators"`
}

type ldapProviderFactory struct {
}

func (g *ldapProviderFactory) Type() string {
	return "GitHubIdentityProvider"
}

func (g *ldapProviderFactory) Create(opts options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var github github
	if err := mapstructure.Decode(opts, &github); err != nil {
		return nil, err
	}

	if github.Endpoint.AuthURL == "" {
		github.Endpoint.AuthURL = authURL
	}
	if github.Endpoint.TokenURL == "" {
		github.Endpoint.TokenURL = tokenURL
	}
	if github.Endpoint.UserInfoURL == "" {
		github.Endpoint.UserInfoURL = userInfoURL
	}
	// fixed options
	opts["endpoint"] = options.DynamicOptions{
		"authURL":     github.Endpoint.AuthURL,
		"tokenURL":    github.Endpoint.TokenURL,
		"userInfoURL": github.Endpoint.UserInfoURL,
	}
	github.Config = &oauth2.Config{
		ClientID:     github.ClientID,
		ClientSecret: github.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  github.Endpoint.AuthURL,
			TokenURL: github.Endpoint.TokenURL,
		},
		RedirectURL: github.RedirectURL,
		Scopes:      github.Scopes,
	}
	return &github, nil
}

func (g githubIdentity) GetUserID() string {
	return g.Login
}

func (g githubIdentity) GetUsername() string {
	return g.Login
}

func (g githubIdentity) GetEmail() string {
	return g.Email
}

func (g *github) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
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

	var githubIdentity githubIdentity
	err = json.Unmarshal(data, &githubIdentity)
	if err != nil {
		return nil, err
	}

	return githubIdentity, nil
}

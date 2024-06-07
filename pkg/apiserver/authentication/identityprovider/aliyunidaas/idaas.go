/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package aliyunidaas

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"golang.org/x/oauth2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

func init() {
	identityprovider.RegisterOAuthProviderFactory(&idaasProviderFactory{})
}

type aliyunIDaaS struct {
	// ClientID is the application's ID.
	ClientID string `json:"clientID" yaml:"clientID"`

	// ClientSecret is the application's secret.
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`

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

	Config *oauth2.Config `json:"-" yaml:"-"`
}

// endpoint represents an OAuth 2.0 provider's authorization and token
// endpoint URLs.
type endpoint struct {
	AuthURL     string `json:"authURL" yaml:"authURL"`
	TokenURL    string `json:"tokenURL" yaml:"tokenURL"`
	UserInfoURL string `json:"userInfoURL" yaml:"userInfoURL"`
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

func (f *idaasProviderFactory) Type() string {
	return "AliyunIDaaSProvider"
}

func (f *idaasProviderFactory) Create(opts options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var idaas aliyunIDaaS
	if err := mapstructure.Decode(opts, &idaas); err != nil {
		return nil, err
	}
	idaas.Config = &oauth2.Config{
		ClientID:     idaas.ClientID,
		ClientSecret: idaas.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:   idaas.Endpoint.AuthURL,
			TokenURL:  idaas.Endpoint.TokenURL,
			AuthStyle: oauth2.AuthStyleAutoDetect,
		},
		RedirectURL: idaas.RedirectURL,
		Scopes:      idaas.Scopes,
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

func (a *aliyunIDaaS) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	// OAuth2 callback, see also https://tools.ietf.org/html/rfc6749#section-4.1.2
	code := req.URL.Query().Get("code")
	ctx := req.Context()
	token, err := a.Config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	resp, err := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token)).Get(a.Endpoint.UserInfoURL)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
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

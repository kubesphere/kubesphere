package casdoor

import (
	auth "github.com/casdoor/casdoor-go-sdk/auth"
	"github.com/mitchellh/mapstructure"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"net/http"
)

func init() {
	identityprovider.RegisterOAuthProvider(&CasdoorOAuthProviderFactory{})
}

type CasdoorIdentity struct {
	auth.User
}

func (c CasdoorIdentity) GetUserID() string {
	return c.Id
}

func (c CasdoorIdentity) GetUsername() string {
	return c.Name
}

func (c CasdoorIdentity) GetEmail() string {
	return c.Email
}

type CasdoorOAuthProvider struct {
	CasdoorEndpoint  string `json:"casdoorEndPoint"`
	ClientID         string `json:"clientID"`
	ClientSecret     string `json:"clientSecret"`
	OrganizationName string `json:"organizationName"`
	ApplicationName  string `json:"applicationName"`
	JwtSecret        string `json:"jwtSecret"`
}

func (c *CasdoorOAuthProvider) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	auth.InitConfig(c.CasdoorEndpoint, c.ClientID, c.ClientSecret, c.JwtSecret, c.OrganizationName, c.ApplicationName)
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")
	token, err := auth.GetOAuthToken(code, state)
	if err != nil {
		return nil, err
	}
	claims, err := auth.ParseJwtToken(token.AccessToken)
	if err != nil {
		return nil, err
	}
	var identity CasdoorIdentity
	identity.User = claims.User
	return identity, nil
}

type CasdoorOAuthProviderFactory struct{}

func (c *CasdoorOAuthProviderFactory) Type() string {
	return "CasdoorOAuthProvider"
}

func (c *CasdoorOAuthProviderFactory) Create(options oauth.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var provider CasdoorOAuthProvider
	if err := mapstructure.Decode(options, &provider); err != nil {
		return nil, err
	}
	return &provider, nil
}

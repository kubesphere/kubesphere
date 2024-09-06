/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cas

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
	gocas "gopkg.in/cas.v2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/server/options"
)

func init() {
	identityprovider.RegisterOAuthProviderFactory(&casProviderFactory{})
}

type cas struct {
	RedirectURL        string `json:"redirectURL" yaml:"redirectURL"`
	CASServerURL       string `json:"casServerURL" yaml:"casServerURL"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify" yaml:"insecureSkipVerify"`
	client             *gocas.RestClient
}

type casProviderFactory struct {
}

type casIdentity struct {
	User string `json:"user"`
}

func (c casIdentity) GetUserID() string {
	return c.User
}

func (c casIdentity) GetUsername() string {
	return c.User
}

func (c casIdentity) GetEmail() string {
	return ""
}

func (f casProviderFactory) Type() string {
	return "CASIdentityProvider"
}

func (f casProviderFactory) Create(opts options.DynamicOptions) (identityprovider.OAuthProvider, error) {
	var cas cas
	if err := mapstructure.Decode(opts, &cas); err != nil {
		return nil, err
	}
	casURL, err := url.Parse(cas.CASServerURL)
	if err != nil {
		return nil, err
	}
	redirectURL, err := url.Parse(cas.RedirectURL)
	if err != nil {
		return nil, err
	}
	cas.client = gocas.NewRestClient(&gocas.RestOptions{
		CasURL:     casURL,
		ServiceURL: redirectURL,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: cas.InsecureSkipVerify},
			},
		},
		URLScheme: nil,
	})
	return &cas, nil
}

func (c cas) IdentityExchangeCallback(req *http.Request) (identityprovider.Identity, error) {
	// CAS callback, see also https://apereo.github.io/cas/6.3.x/protocol/CAS-Protocol-V2-Specification.html#25-servicevalidate-cas-20
	ticket := req.URL.Query().Get("ticket")
	resp, err := c.client.ValidateServiceTicket(gocas.ServiceTicket(ticket))
	if err != nil {
		return nil, fmt.Errorf("cas: failed to validate service ticket : %v", err)
	}
	return &casIdentity{User: resp.User}, nil
}

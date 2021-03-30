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
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

func TestDefaultAuthOptions(t *testing.T) {
	oneDay := time.Second * 86400
	zero := time.Duration(0)
	expect := Client{
		Name:                         "default",
		RespondWithChallenges:        true,
		Secret:                       "kubesphere",
		RedirectURIs:                 []string{AllowAllRedirectURI},
		GrantMethod:                  GrantHandlerAuto,
		ScopeRestrictions:            []string{"full"},
		AccessTokenMaxAge:            &oneDay,
		AccessTokenInactivityTimeout: &zero,
	}

	options := NewOptions()
	client, err := options.OAuthClient("default")
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expect, client); len(diff) != 0 {
		t.Errorf("%T differ (-got, +expected), %s", expect, diff)
	}
}

func TestClientResolveRedirectURL(t *testing.T) {
	options := NewOptions()
	defaultClient, err := options.OAuthClient("default")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		Name        string
		client      Client
		expectError error
		expectURL   string
	}{
		{
			Name:        "default client test",
			client:      defaultClient,
			expectError: nil,
			expectURL:   "https://localhost:8080/auth/cb",
		},
		{
			Name: "custom client test",
			client: Client{
				Name:                  "default",
				RespondWithChallenges: true,
				RedirectURIs:          []string{"https://foo.bar.com/oauth/cb"},
				GrantMethod:           GrantHandlerAuto,
				ScopeRestrictions:     []string{"full"},
			},
			expectError: ErrorRedirectURLNotAllowed,
			expectURL:   "https://foo.bar.com/oauth/err",
		},
		{
			Name: "custom client test",
			client: Client{
				Name:                  "default",
				RespondWithChallenges: true,
				RedirectURIs:          []string{AllowAllRedirectURI, "https://foo.bar.com/oauth/cb"},
				GrantMethod:           GrantHandlerAuto,
				ScopeRestrictions:     []string{"full"},
			},
			expectError: nil,
			expectURL:   "https://foo.bar.com/oauth/err2",
		},
	}

	for _, test := range tests {
		redirectURL, err := test.client.ResolveRedirectURL(test.expectURL)
		if err != test.expectError {
			t.Errorf("expected error: %s, got: %s", test.expectError, err)
		}
		if test.expectError == nil && test.expectURL != redirectURL {
			t.Errorf("expected redirect url: %s, got: %s", test.expectURL, redirectURL)
		}
	}
}

func TestDynamicOptions_MarshalJSON(t *testing.T) {
	config := `
accessTokenMaxAge: 1h
accessTokenInactivityTimeout: 30m
identityProviders:
  - name: ldap
    type: LDAPIdentityProvider
    mappingMethod: auto
    provider:
      host: xxxx.sn.mynetname.net:389
      managerDN: uid=root,cn=users,dc=xxxx,dc=sn,dc=mynetname,dc=net
      managerPassword: xxxx
      userSearchBase: dc=xxxx,dc=sn,dc=mynetname,dc=net
      loginAttribute: uid
      mailAttribute: mail
  - name: github
    type: GitHubIdentityProvider
    mappingMethod: mixed
    provider:
      clientID: 'xxxxxx'
      clientSecret: 'xxxxxx'
      endpoint:
        authURL: 'https://github.com/login/oauth/authorize'
        tokenURL: 'https://github.com/login/oauth/access_token'
      redirectURL: 'https://ks-console/oauth/redirect'
      scopes:
      - user
`
	var options Options
	if err := yaml.Unmarshal([]byte(config), &options); err != nil {
		t.Error(err)
	}
	expected := `{"identityProviders":[{"name":"ldap","mappingMethod":"auto","type":"LDAPIdentityProvider","provider":{"host":"xxxx.sn.mynetname.net:389","loginAttribute":"uid","mailAttribute":"mail","managerDN":"uid=root,cn=users,dc=xxxx,dc=sn,dc=mynetname,dc=net","userSearchBase":"dc=xxxx,dc=sn,dc=mynetname,dc=net"}},{"name":"github","mappingMethod":"mixed","type":"GitHubIdentityProvider","provider":{"clientID":"xxxxxx","endpoint":{"authURL":"https://github.com/login/oauth/authorize","tokenURL":"https://github.com/login/oauth/access_token"},"redirectURL":"https://ks-console/oauth/redirect","scopes":["user"]}}],"accessTokenMaxAge":3600000000000,"accessTokenInactivityTimeout":1800000000000}`
	output, _ := json.Marshal(options)
	if expected != string(output) {
		t.Errorf("expected: %s, but got: %s", expected, output)
	}
}

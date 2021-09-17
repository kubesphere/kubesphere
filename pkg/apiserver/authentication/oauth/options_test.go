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

	"gopkg.in/yaml.v3"
)

func TestClientResolveRedirectURL(t *testing.T) {

	tests := []struct {
		Name      string
		client    Client
		wantErr   bool
		expectURL string
	}{
		{
			Name: "custom client test",
			client: Client{
				Name:                  "custom",
				RespondWithChallenges: true,
				RedirectURIs:          []string{AllowAllRedirectURI, "https://foo.bar.com/oauth/cb"},
				GrantMethod:           GrantHandlerAuto,
			},
			wantErr:   false,
			expectURL: "https://foo.bar.com/oauth/cb",
		},
		{
			Name: "custom client test",
			client: Client{
				Name:                  "custom",
				RespondWithChallenges: true,
				RedirectURIs:          []string{"https://foo.bar.com/oauth/cb"},
				GrantMethod:           GrantHandlerAuto,
			},
			wantErr:   true,
			expectURL: "https://foo.bar.com/oauth/cb2",
		},
	}

	for _, test := range tests {
		redirectURL, err := test.client.ResolveRedirectURL(test.expectURL)
		if (err != nil) != test.wantErr {
			t.Errorf("ResolveRedirectURL() error = %+v, wantErr %+v", err, test.wantErr)
			return
		}
		if redirectURL != nil && test.expectURL != redirectURL.String() {
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
	expected := `{"identityProviders":[{"name":"ldap","mappingMethod":"auto","disableLoginConfirmation":false,"type":"LDAPIdentityProvider","provider":{"host":"xxxx.sn.mynetname.net:389","loginAttribute":"uid","mailAttribute":"mail","managerDN":"uid=root,cn=users,dc=xxxx,dc=sn,dc=mynetname,dc=net","userSearchBase":"dc=xxxx,dc=sn,dc=mynetname,dc=net"}},{"name":"github","mappingMethod":"mixed","disableLoginConfirmation":false,"type":"GitHubIdentityProvider","provider":{"clientID":"xxxxxx","endpoint":{"authURL":"https://github.com/login/oauth/authorize","tokenURL":"https://github.com/login/oauth/access_token"},"redirectURL":"https://ks-console/oauth/redirect","scopes":["user"]}}],"accessTokenMaxAge":3600000000000,"accessTokenInactivityTimeout":1800000000000}`
	output, _ := json.Marshal(options)
	if expected != string(output) {
		t.Errorf("expected: %s, but got: %s", expected, output)
	}
}

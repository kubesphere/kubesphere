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

package github

import (
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var githubServer *httptest.Server

func TestGithub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitHub Identity Provider Suite")
}

var _ = BeforeSuite(func(done Done) {
	githubServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data map[string]interface{}
		switch r.RequestURI {
		case "/login/oauth/access_token":
			data = map[string]interface{}{
				"access_token": "e72e16c7e42f292c6912e7710c838347ae178b4a",
				"scope":        "user,repo,gist",
				"token_type":   "bearer",
			}
		case "/user":
			data = map[string]interface{}{
				"login": "test",
				"email": "test@kubesphere.io",
			}
		default:
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("not implemented"))
			return
		}

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	}))
	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	githubServer.Close()
})

var _ = Describe("GitHub", func() {
	Context("GitHub", func() {
		var (
			provider identityprovider.OAuthProvider
			err      error
		)
		It("should configure successfully", func() {
			configYAML := `
clientID: de6ff8bed0304e487b6e
clientSecret: 2b70536f79ec8d2939863509d05e2a71c268b9af
redirectURL: "http://ks-console/oauth/redirect"
scopes:
- user
`
			config := mustUnmarshalYAML(configYAML)
			factory := ldapProviderFactory{}
			provider, err = factory.Create(config)
			Expect(err).Should(BeNil())
			expected := &github{
				ClientID:     "de6ff8bed0304e487b6e",
				ClientSecret: "2b70536f79ec8d2939863509d05e2a71c268b9af",
				Endpoint: endpoint{
					AuthURL:     authURL,
					TokenURL:    tokenURL,
					UserInfoURL: userInfoURL,
				},
				RedirectURL: "http://ks-console/oauth/redirect",
				Scopes:      []string{"user"},
				Config: &oauth2.Config{
					ClientID:     "de6ff8bed0304e487b6e",
					ClientSecret: "2b70536f79ec8d2939863509d05e2a71c268b9af",
					Endpoint: oauth2.Endpoint{
						AuthURL:  authURL,
						TokenURL: tokenURL,
					},
					RedirectURL: "http://ks-console/oauth/redirect",
					Scopes:      []string{"user"},
				},
			}
			Expect(provider).Should(Equal(expected))
		})
		It("should configure successfully", func() {
			config := oauth.DynamicOptions{
				"clientID":           "de6ff8bed0304e487b6e",
				"clientSecret":       "2b70536f79ec8d2939863509d05e2a71c268b9af",
				"redirectURL":        "http://ks-console/oauth/redirect",
				"insecureSkipVerify": true,
				"endpoint": oauth.DynamicOptions{
					"authURL":     fmt.Sprintf("%s/login/oauth/authorize", githubServer.URL),
					"tokenURL":    fmt.Sprintf("%s/login/oauth/access_token", githubServer.URL),
					"userInfoURL": fmt.Sprintf("%s/user", githubServer.URL),
				},
			}
			factory := ldapProviderFactory{}
			provider, err = factory.Create(config)
			Expect(err).Should(BeNil())
			expected := oauth.DynamicOptions{
				"clientID":           "de6ff8bed0304e487b6e",
				"clientSecret":       "2b70536f79ec8d2939863509d05e2a71c268b9af",
				"redirectURL":        "http://ks-console/oauth/redirect",
				"insecureSkipVerify": true,
				"endpoint": oauth.DynamicOptions{
					"authURL":     fmt.Sprintf("%s/login/oauth/authorize", githubServer.URL),
					"tokenURL":    fmt.Sprintf("%s/login/oauth/access_token", githubServer.URL),
					"userInfoURL": fmt.Sprintf("%s/user", githubServer.URL),
				},
			}
			Expect(config).Should(Equal(expected))
		})
		It("should login successfully", func() {
			identity, err := provider.IdentityExchange("3389")
			Expect(err).Should(BeNil())
			Expect(identity.GetUserID()).Should(Equal("test"))
			Expect(identity.GetUsername()).Should(Equal("test"))
			Expect(identity.GetEmail()).Should(Equal("test@kubesphere.io"))
		})
	})
})

func mustUnmarshalYAML(data string) oauth.DynamicOptions {
	var dynamicOptions oauth.DynamicOptions
	_ = yaml.Unmarshal([]byte(data), &dynamicOptions)
	return dynamicOptions
}

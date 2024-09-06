/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"kubesphere.io/kubesphere/pkg/server/options"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

var githubServer *httptest.Server

func TestGithub(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GitHub Identity Provider Suite")
}

var _ = BeforeSuite(func() {
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
})

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
redirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/github"
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
				RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/github",
				Scopes:      []string{"user"},
				Config: &oauth2.Config{
					ClientID:     "de6ff8bed0304e487b6e",
					ClientSecret: "2b70536f79ec8d2939863509d05e2a71c268b9af",
					Endpoint: oauth2.Endpoint{
						AuthURL:  authURL,
						TokenURL: tokenURL,
					},
					RedirectURL: "https://ks-console.kubesphere-system.svc/oauth/redirect/github",
					Scopes:      []string{"user"},
				},
			}
			Expect(provider).Should(Equal(expected))
		})
		It("should configure successfully", func() {
			config := options.DynamicOptions{
				"clientID":           "de6ff8bed0304e487b6e",
				"clientSecret":       "2b70536f79ec8d2939863509d05e2a71c268b9af",
				"redirectURL":        "https://ks-console.kubesphere-system.svc/oauth/redirect/github",
				"insecureSkipVerify": true,
				"endpoint": options.DynamicOptions{
					"authURL":     fmt.Sprintf("%s/login/oauth/authorize", githubServer.URL),
					"tokenURL":    fmt.Sprintf("%s/login/oauth/access_token", githubServer.URL),
					"userInfoURL": fmt.Sprintf("%s/user", githubServer.URL),
				},
			}
			factory := ldapProviderFactory{}
			provider, err = factory.Create(config)
			Expect(err).Should(BeNil())
			expected := options.DynamicOptions{
				"clientID":           "de6ff8bed0304e487b6e",
				"clientSecret":       "2b70536f79ec8d2939863509d05e2a71c268b9af",
				"redirectURL":        "https://ks-console.kubesphere-system.svc/oauth/redirect/github",
				"insecureSkipVerify": true,
				"endpoint": options.DynamicOptions{
					"authURL":     fmt.Sprintf("%s/login/oauth/authorize", githubServer.URL),
					"tokenURL":    fmt.Sprintf("%s/login/oauth/access_token", githubServer.URL),
					"userInfoURL": fmt.Sprintf("%s/user", githubServer.URL),
				},
			}
			Expect(config).Should(Equal(expected))
		})
		It("should login successfully", func() {
			url, _ := url.Parse("https://ks-console.kubesphere-system.svc/oauth/redirect/test?code=00000")
			req := &http.Request{URL: url}
			identity, err := provider.IdentityExchangeCallback(req)
			Expect(err).Should(BeNil())
			Expect(identity.GetUserID()).Should(Equal("test"))
			Expect(identity.GetUsername()).Should(Equal("test"))
			Expect(identity.GetEmail()).Should(Equal("test@kubesphere.io"))
		})
	})
})

func mustUnmarshalYAML(data string) options.DynamicOptions {
	var dynamicOptions options.DynamicOptions
	_ = yaml.Unmarshal([]byte(data), &dynamicOptions)
	return dynamicOptions
}

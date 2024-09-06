/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package oidc

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"kubesphere.io/kubesphere/pkg/server/options"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"gopkg.in/square/go-jose.v2"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/identityprovider"
)

var (
	oidcServer *httptest.Server
)

func TestOIDC(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OIDC Identity Provider Suite")
}

var _ = BeforeSuite(func() {
	privateKey, err := rsa.GenerateKey(cryptorand.Reader, 2048)
	Expect(err).Should(BeNil())
	jwk := jose.JSONWebKey{
		Key:       privateKey,
		KeyID:     "keyID",
		Algorithm: "RSA",
	}
	oidcServer = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data interface{}
		switch r.RequestURI {
		case "/.well-known/openid-configuration":
			data = map[string]interface{}{
				"issuer":                 oidcServer.URL,
				"token_endpoint":         fmt.Sprintf("%s/token", oidcServer.URL),
				"authorization_endpoint": fmt.Sprintf("%s/authorize", oidcServer.URL),
				"userinfo_endpoint":      fmt.Sprintf("%s/userinfo", oidcServer.URL),
				"end_session_endpoint":   fmt.Sprintf("%s/endsession", oidcServer.URL),
				"jwks_uri":               fmt.Sprintf("%s/keys", oidcServer.URL),
				"response_types_supported": []string{
					"code",
					"token",
					"id_token",
					"none",
				},
				"id_token_signing_alg_values_supported": []string{
					"RS256",
				},
				"scopes_supported": []string{
					"openid",
					"email",
					"profile",
				},
				"token_endpoint_auth_methods_supported": []string{
					"client_secret_post",
					"client_secret_basic",
				},
				"claims_supported": []string{
					"aud",
					"email",
					"email_verified",
					"exp",
					"iat",
					"iss",
					"name",
					"sub",
				},
				"code_challenge_methods_supported": []string{
					"plain",
					"S256",
				},
				"grant_types_supported": []string{
					"authorization_code",
					"refresh_token",
				},
			}
		case "/user":
			data = map[string]interface{}{
				"login": "test",
				"email": "test@kubesphere.io",
			}
		case "/keys":
			data = map[string]interface{}{
				"keys": []map[string]interface{}{{
					"alg": jwk.Algorithm,
					"kty": jwk.Algorithm,
					"kid": jwk.KeyID,
					"n":   n(&privateKey.PublicKey),
					"e":   e(&privateKey.PublicKey),
				}},
			}
		case "/token":
			claims := jwt.MapClaims{
				"iss":            oidcServer.URL,
				"sub":            "110169484474386276334",
				"aud":            "kubesphere",
				"email":          "test@kubesphere.io",
				"email_verified": "true",
				"name":           "test",
				"iat":            time.Now().Unix(),
				"exp":            time.Now().Add(10 * time.Hour).Unix(),
			}
			idToken, _ := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
			data = map[string]interface{}{
				"access_token": "e72e16c7e42f292c6912e7710c838347ae178b4a",
				"id_token":     idToken,
				"token_type":   "Bearer",
				"expires_in":   3600,
			}
		default:
			fmt.Println(r.URL)
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
	oidcServer.Close()
})

var _ = Describe("OIDC", func() {
	Context("OIDC", func() {
		var (
			provider identityprovider.OAuthProvider
			err      error
		)
		It("should configure successfully", func() {
			config := options.DynamicOptions{
				"issuer":             oidcServer.URL,
				"clientID":           "kubesphere",
				"clientSecret":       "c53e80ab92d48ab12f4e7f1f6976d1bdc996e0d7",
				"redirectURL":        "https://ks-console.kubesphere-system.svc/oauth/redirect/oidc",
				"insecureSkipVerify": true,
			}
			factory := oidcProviderFactory{}
			provider, err = factory.Create(config)
			Expect(err).Should(BeNil())
			expected := options.DynamicOptions{
				"issuer":             oidcServer.URL,
				"clientID":           "kubesphere",
				"clientSecret":       "c53e80ab92d48ab12f4e7f1f6976d1bdc996e0d7",
				"redirectURL":        "https://ks-console.kubesphere-system.svc/oauth/redirect/oidc",
				"insecureSkipVerify": true,
				"endpoint": options.DynamicOptions{
					"authURL":       fmt.Sprintf("%s/authorize", oidcServer.URL),
					"tokenURL":      fmt.Sprintf("%s/token", oidcServer.URL),
					"userInfoURL":   fmt.Sprintf("%s/userinfo", oidcServer.URL),
					"jwksURL":       fmt.Sprintf("%s/keys", oidcServer.URL),
					"endSessionURL": fmt.Sprintf("%s/endsession", oidcServer.URL),
				},
			}
			Expect(config).Should(Equal(expected))
		})
		It("should login successfully", func() {
			url, _ := url.Parse("https://ks-console.kubesphere-system.svc/oauth/redirect/oidc?code=00000")
			req := &http.Request{URL: url}
			identity, err := provider.IdentityExchangeCallback(req)
			Expect(err).Should(BeNil())
			Expect(identity.GetUserID()).Should(Equal("110169484474386276334"))
			Expect(identity.GetUsername()).Should(Equal("test"))
			Expect(identity.GetEmail()).Should(Equal("test@kubesphere.io"))
		})
	})
})

func n(pub *rsa.PublicKey) string {
	return encode(pub.N.Bytes())
}

func e(pub *rsa.PublicKey) string {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(pub.E))
	return encode(bytes.TrimLeft(data, "\x00"))
}

func encode(payload []byte) string {
	result := base64.URLEncoding.EncodeToString(payload)
	return strings.TrimRight(result, "=")
}

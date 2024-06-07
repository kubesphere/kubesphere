/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-containerregistry/pkg/authn"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildSecret(registry, username, password string, insecure bool) *v1.Secret {
	auth := fmt.Sprintf("%s:%s", username, password)
	authString := fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"email\":\"\",\"auth\":\"%s\"}}}", registry, username, password, base64.StdEncoding.EncodeToString([]byte(auth)))

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "docker",
			Namespace: v1.NamespaceDefault,
		},
		Data: map[string][]byte{
			v1.DockerConfigJsonKey: []byte(authString),
		},
		Type: v1.SecretTypeDockerConfigJson,
	}

	if insecure {
		secret.Annotations = make(map[string]string)
		secret.Annotations[forceInsecure] = "true"
	}

	return secret
}

func TestSecretAuthenticator(t *testing.T) {
	secret := buildSecret("dockerhub.qingcloud.com", "guest", "guest", false)

	secretAuthenticator, err := NewSecretAuthenticator(secret)
	if err != nil {
		t.Fatal(err)
	}

	auth, err := secretAuthenticator.Authorization()
	if err != nil {
		t.Fatal(err)
	}

	expected := &authn.AuthConfig{
		Username: "guest",
		Password: "guest",
		Auth:     "Z3Vlc3Q6Z3Vlc3Q=",
	}

	if diff := cmp.Diff(auth, expected); len(diff) != 0 {
		t.Errorf("%T, got+ expected-, %s", expected, diff)
	}
}

func TestAuthn(t *testing.T) {
	testCases := []struct {
		name      string
		secret    *v1.Secret
		auth      bool
		expectErr bool
	}{
		{
			name:      "Should authenticate with correct credential",
			secret:    buildSecret("https://dockerhub.qingcloud.com", "guest", "guest", false),
			auth:      true,
			expectErr: false,
		},
		{
			name:      "Shouldn't authenticate with incorrect credentials",
			secret:    buildSecret("https://index.docker.io", "foo", "bar", false),
			auth:      false,
			expectErr: true,
		},
		{
			name:      "Shouldn't authenticate with no credentials",
			secret:    nil,
			auth:      false,
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			secretAuthenticator, err := NewSecretAuthenticator(testCase.secret)
			if err != nil {
				t.Errorf("error creating secretAuthenticator, %v", err)
			}

			ok, err := secretAuthenticator.Auth()
			if testCase.auth != ok {
				t.Errorf("expected auth result: %v, but got %v", testCase.auth, ok)
			}

			if testCase.expectErr && err == nil {
				t.Errorf("expected error, but got nil")
			}

			if !testCase.expectErr && err != nil {
				t.Errorf("authentication error, %v", err)
			}
		})
	}
}

var (
	registryServer    *httptest.Server
	tlsRegistryServer *httptest.Server
)

func TestRegistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registry Test Suite")
}

var _ = BeforeSuite(func() {
	// anonymous registry
	fakeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	tlsRegistryServer = httptest.NewTLSServer(fakeHandler)
	registryServer = httptest.NewServer(fakeHandler)
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	gexec.KillAndWait(5 * time.Second)
	registryServer.Close()
	tlsRegistryServer.Close()
})

var _ = Describe("Registry", func() {
	Context("Registry", func() {
		It("skip TLS certification checks", func() {
			secret := buildSecret(tlsRegistryServer.URL, "", "", true)
			secretAuthenticator, err := NewSecretAuthenticator(secret)
			Expect(err).Should(BeNil())
			_, err = secretAuthenticator.Auth()
			Expect(err).Should(BeNil())
		})
		It("self-signed certs are not trusted", func() {
			secret := buildSecret(tlsRegistryServer.URL, "", "", false)
			secretAuthenticator, err := NewSecretAuthenticator(secret)
			Expect(err).Should(BeNil())
			_, err = secretAuthenticator.Auth()
			Expect(err).ShouldNot(BeNil())
		})
		It("insecure registry", func() {
			// Loopback addr always be insecure, http scheme will be used.
			secret := buildSecret(registryServer.URL, "", "", false)
			secretAuthenticator, err := NewSecretAuthenticator(secret)
			Expect(err).Should(BeNil())
			_, err = secretAuthenticator.Auth()
			Expect(err).Should(BeNil())
		})
	})
})

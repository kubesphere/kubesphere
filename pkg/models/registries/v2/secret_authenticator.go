/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package v2

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	v1 "k8s.io/api/core/v1"
)

const (
	forceInsecure = "secret.kubesphere.io/force-insecure"
)

type SecretAuthenticator interface {
	Options() []Option

	Auth() (bool, error)

	Authorization() (*authn.AuthConfig, error)
}

type secretAuthenticator struct {
	auths    DockerConfig
	insecure bool // force using insecure when talk to the remote registry, even registry address starts with https
}

func NewSecretAuthenticator(secret *v1.Secret) (SecretAuthenticator, error) {

	if secret == nil {
		return &secretAuthenticator{}, nil
	}

	sa := &secretAuthenticator{
		insecure: false,
	}

	if secret.Type != v1.SecretTypeDockerConfigJson {
		return nil, fmt.Errorf("expected secret type: %s, got: %s", v1.SecretTypeDockerConfigJson, secret.Type)
	}

	// force insecure if secret has annotation forceInsecure
	if val, ok := secret.Annotations[forceInsecure]; ok && val == "true" {
		sa.insecure = true
	}

	configJson, ok := secret.Data[v1.DockerConfigJsonKey]
	if !ok {
		return nil, fmt.Errorf("expected key %s in data, found none", v1.DockerConfigJsonKey)
	}

	dockerConfigJSON := DockerConfigJSON{}
	if err := json.Unmarshal(configJson, &dockerConfigJSON); err != nil {
		return nil, err
	}

	if len(dockerConfigJSON.Auths) == 0 {
		return nil, fmt.Errorf("not found valid auth in secret, %v", dockerConfigJSON)
	}

	sa.auths = dockerConfigJSON.Auths

	return sa, nil
}

func (s *secretAuthenticator) Authorization() (*authn.AuthConfig, error) {
	for _, v := range s.auths {
		return &authn.AuthConfig{
			Username: v.Username,
			Password: v.Password,
			Auth:     v.Auth,
		}, nil
	}
	return &authn.AuthConfig{}, nil
}

func (s *secretAuthenticator) Auth() (bool, error) {
	for k := range s.auths {
		return s.AuthRegistry(k)
	}
	return false, fmt.Errorf("no registry found in secret")
}

func (s *secretAuthenticator) AuthRegistry(reg string) (bool, error) {
	url, err := url.Parse(reg) // in case reg is unformatted like http://docker.index.io
	if err != nil {
		return false, err
	}

	options := make([]name.Option, 0)
	if url.Scheme == "http" {
		// allows image references to be fetched without TLS
		// transport.NewWithContext will auto-select the right scheme
		options = append(options, name.Insecure)
	}
	tr := http.DefaultTransport.(*http.Transport).Clone()
	// skip tls verify
	if s.insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	registry, err := name.NewRegistry(url.Host, options...)
	if err != nil {
		return false, err
	}

	ctx := context.TODO()
	_, err = transport.NewWithContext(ctx, registry, s, tr, []string{})
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *secretAuthenticator) Options() []Option {
	options := make([]Option, 0)
	options = append(options, WithAuth(s))
	if s.registryScheme() == "http" {
		options = append(options, Insecure)
	}
	if s.insecure {
		tr := http.DefaultTransport.(*http.Transport).Clone()
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		options = append(options, WithTransport(tr))
	}
	return options
}

func (s *secretAuthenticator) registryScheme() string {
	for registry := range s.auths {
		u, err := url.Parse(registry)
		if err == nil {
			return u.Scheme
		}
	}
	return "https"
}

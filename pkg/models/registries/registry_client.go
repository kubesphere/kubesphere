/*
Copyright 2020 KubeSphere Authors

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

package registries

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"io"
	"io/ioutil"
	log "k8s.io/klog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultDockerRegistry is the default docker registry address.
	DefaultDockerRegistry = "https://registry-1.docker.io"

	DefaultDockerHub = "docker.io"

	DefaultTimeout = 30 * time.Second
)

var (
	bearerRegex = regexp.MustCompile(
		`^\s*Bearer\s+(.*)$`)
	basicRegex = regexp.MustCompile(`^\s*Basic\s+.*$`)

	// ErrBasicAuth indicates that the repository requires basic rather than token authentication.
	ErrBasicAuth = errors.New("basic auth required")

	gcrMatcher = regexp.MustCompile(`https://([a-z]+\.|)gcr\.io/`)
)

// Registry defines the client for retrieving information from the registry API.
type Registry struct {
	URL      string
	Domain   string
	Username string
	Password string
	Client   *http.Client
	Opt      RegistryOpt
}

// Opt holds the options for a new registry.
type RegistryOpt struct {
	Domain  string
	Timeout time.Duration
	Headers map[string]string
	UseSSL  bool
}

type authToken struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

type authService struct {
	Realm   *url.URL
	Service string
	Scope   []string
}

func CreateRegistryClient(username, password, domain string, useSSL bool) (*Registry, error) {
	authDomain := domain
	auth, err := GetAuthConfig(username, password, authDomain)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	// Create the registry client.
	return New(auth, RegistryOpt{
		Domain: domain,
		UseSSL: useSSL,
	})
}

// GetAuthConfig returns the docker registry AuthConfig.
func GetAuthConfig(username, password, registry string) (types.AuthConfig, error) {
	registry = setDefaultRegistry(registry)
	if username != "" && password != "" {
		return types.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: registry,
		}, nil
	}

	return types.AuthConfig{
		ServerAddress: registry,
	}, nil

}

func setDefaultRegistry(serverAddress string) string {
	if serverAddress == DefaultDockerHub || serverAddress == "" {
		serverAddress = DefaultDockerRegistry
	}

	return serverAddress
}

func newFromTransport(auth types.AuthConfig, opt RegistryOpt) (*Registry, error) {
	if len(opt.Domain) < 1 || opt.Domain == DefaultDockerHub {
		opt.Domain = auth.ServerAddress
	}
	registryUrl := strings.TrimSuffix(opt.Domain, "/")

	if !strings.HasPrefix(registryUrl, "http://") && !strings.HasPrefix(registryUrl, "https://") {
		if opt.UseSSL {
			registryUrl = "https://" + registryUrl
		} else {
			registryUrl = "http://" + registryUrl
		}
	}

	registryURL, _ := url.Parse(registryUrl)
	registry := &Registry{
		URL:    registryURL.String(),
		Domain: registryURL.Host,
		Client: &http.Client{
			Timeout: DefaultTimeout,
		},
		Username: auth.Username,
		Password: auth.Password,
		Opt:      opt,
	}

	return registry, nil
}

// url returns a registry URL with the passed arguments concatenated.
func (r *Registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

// New creates a new Registry struct with the given URL and credentials.
func New(auth types.AuthConfig, opt RegistryOpt) (*Registry, error) {

	return newFromTransport(auth, opt)
}

// Decompress response.body.
func GetRespBody(resp *http.Response) ([]byte, error) {
	var reader io.ReadCloser
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, _ = gzip.NewReader(resp.Body)
	} else {
		reader = resp.Body
	}
	resBody, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return resBody, err
}

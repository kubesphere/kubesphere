package registries

import (
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	log "github.com/golang/glog"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultDockerRegistry is the default docker registry address.
	DefaultDockerRegistry = "https://registry-1.docker.io"

	DefaultTimeout = 30 * time.Second
)

var (
	bearerRegex = regexp.MustCompile(
		`^\s*Bearer\s+(.*)$`)
	basicRegex = regexp.MustCompile(`^\s*Basic\s+.*$`)

	// ErrBasicAuth indicates that the repository requires basic rather than token authentication.
	ErrBasicAuth = errors.New("basic auth required")

	gcrMatcher = regexp.MustCompile(`https://([a-z]+\.|)gcr\.io/`)

	reProtocol = regexp.MustCompile("^https?://")
)

// Registry defines the client for retrieving information from the registry API.
type Registry struct {
	URL      string
	Domain   string
	Username string
	Password string
	Client   *http.Client
	Opt      Opt
}

// Opt holds the options for a new registry.
type Opt struct {
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

func CreateRegistryClient(username, password, domain string) (*Registry, error) {
	authDomain := domain
	auth, err := GetAuthConfig(username, password, authDomain)
	if err != nil {
		return nil, err
	}

	// Create the registry client.
	log.Infof("domain: %s", domain)
	log.Infof("server address: %s", auth.ServerAddress)

	return New(auth, Opt{
		Domain: domain,
	})
}

// GetAuthConfig returns the docker registry AuthConfig.
func GetAuthConfig(username, password, registry string) (types.AuthConfig, error) {
	registry = setDefaultRegistry(registry)
	if username != "" && password != "" && registry != "" {
		return types.AuthConfig{
			Username:      username,
			Password:      password,
			ServerAddress: registry,
		}, nil
	}

	log.Info("Using registry ", registry, " with no authentication")

	return types.AuthConfig{
		ServerAddress: registry,
	}, nil

}

func setDefaultRegistry(serverAddress string) string {
	if serverAddress == "docker.io" || serverAddress == "" {
		serverAddress = DefaultDockerRegistry
	}

	return serverAddress
}

func newFromTransport(auth types.AuthConfig, opt Opt) (*Registry, error) {
	if len(opt.Domain) < 1 || opt.Domain == "docker.io" {
		opt.Domain = auth.ServerAddress
	}
	url := strings.TrimSuffix(opt.Domain, "/")
	authURL := strings.TrimSuffix(auth.ServerAddress, "/")

	if !reProtocol.MatchString(url) {
		if opt.UseSSL {
			url = "https://" + url
		} else {
			url = "http://" + url
		}
	}

	if !reProtocol.MatchString(authURL) {
		if opt.UseSSL {
			authURL = "https://" + authURL
		} else {
			authURL = "http://" + authURL
		}
	}

	registry := &Registry{
		URL:    url,
		Domain: reProtocol.ReplaceAllString(url, ""),
		Client: &http.Client{
			Timeout: DefaultTimeout,
		},
		Username: auth.Username,
		Password: auth.Password,
		Opt:      opt,
	}

	return registry, nil
}

// url returns a registry URL with the passed arguements concatenated.
func (r *Registry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

// New creates a new Registry struct with the given URL and credentials.
func New(auth types.AuthConfig, opt Opt) (*Registry, error) {

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

package sonarqube

import (
	"fmt"
	"github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog"
	"strings"
)

type Client struct {
	client *sonargo.Client
}

func NewSonarQubeClient(options *Options) (*Client, error) {
	var endpoint string

	if strings.HasSuffix(options.Host, "/") {
		endpoint = fmt.Sprintf("%sapi/", options.Host)
	} else {
		endpoint = fmt.Sprintf("%s/api/", options.Host)
	}

	sonar, err := sonargo.NewClientWithToken(endpoint, options.Token)
	if err != nil {
		klog.Errorf("failed to connect to sonarqube service, %+v", err)
		return nil, err
	}

	return &Client{client: sonar}, err
}

func NewSonarQubeClientOrDie(options *Options) *Client {
	var endpoint string

	if strings.HasSuffix(options.Host, "/") {
		endpoint = fmt.Sprintf("%sapi/", options.Host)
	} else {
		endpoint = fmt.Sprintf("%s/api/", options.Host)
	}

	sonar, err := sonargo.NewClientWithToken(endpoint, options.Token)
	if err != nil {
		klog.Errorf("failed to connect to sonarqube service, %+v", err)
		panic(err)
	}

	return &Client{client: sonar}
}

// return sonarqube client
// Also we can wrap some methods to avoid direct use sonar client
func (s *Client) SonarQube() *sonargo.Client {
	return s.client
}

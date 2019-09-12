package sonarqube

import (
	"fmt"
	"github.com/kubesphere/sonargo/sonar"
	"k8s.io/klog"
	"strings"
)

type SonarQubeClient struct {
	client *sonargo.Client
}

func NewSonarQubeClient(options *SonarQubeOptions) (*SonarQubeClient, error) {
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

	return &SonarQubeClient{client: sonar}, err
}

func NewSonarQubeClientOrDie(options *SonarQubeOptions) *SonarQubeClient {
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

	return &SonarQubeClient{client: sonar}
}

// return sonarqube client
// Also we can wrap some methods to avoid direct use sonar client
func (s *SonarQubeClient) SonarQube() *sonargo.Client {
	return s.client
}

/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package dockerhub

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/models/registries/imagesearch"
)

const (
	DockerHubRegisterProvider = "DockerHubRegistryProvider"
	dockerHubSearchUrl        = "v2/search/repositories?query=%s"
	dockerHubHost             = "https://hub.docker.com"
)

func init() {
	imagesearch.RegistrySearchProvider(&dockerHubSearchProviderFactory{})
}

var _ imagesearch.SearchProvider = &dockerHubSearchProvider{}

type dockerHubSearchProvider struct {
	HttpClient *http.Client `json:"-" yaml:"-"`
}

type searchResponse struct {
	Results []result `json:"results"`
}

type result struct {
	RepoName string `json:"repo_name"`
}

func (d dockerHubSearchProvider) Search(imageName string, config imagesearch.SearchConfig) (*imagesearch.Results, error) {
	url := fmt.Sprintf("%s/%s", dockerHubHost, fmt.Sprintf(dockerHubSearchUrl, imageName))
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if config.Username != "" {
		authCode := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", config.Username, config.Password))))
		request.Header.Set("Authorization", authCode)
	}

	resp, err := d.HttpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		klog.Errorf("search images failed with status code: %d, %s", resp.StatusCode, string(bytes))
		return nil, fmt.Errorf("search images failed with status code: %d", resp.StatusCode)
	}

	searchResp := &searchResponse{}
	err = json.Unmarshal(bytes, searchResp)
	if err != nil {
		return nil, err
	}
	imageResult := &imagesearch.Results{
		Entries: make([]string, 0),
	}
	for _, v := range searchResp.Results {
		imageResult.Entries = append(imageResult.Entries, v.RepoName)
	}

	imageResult.Total = int64(len(imageResult.Entries))

	return imageResult, nil
}

var _ imagesearch.SearchProviderFactory = &dockerHubSearchProviderFactory{}

type dockerHubSearchProviderFactory struct{}

func (d dockerHubSearchProviderFactory) Type() string {
	return DockerHubRegisterProvider
}

func (d dockerHubSearchProviderFactory) Create(_ map[string]interface{}) (imagesearch.SearchProvider, error) {
	var provider dockerHubSearchProvider
	provider.HttpClient = http.DefaultClient
	return provider, nil
}

/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package harbor

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/models/registries/imagesearch"
)

const (
	HarborRegisterProvider = "HarborRegistryProvider"
	harborSearchUrl        = "api/v2.0/search?q=%s"
)

func init() {
	imagesearch.RegistrySearchProvider(&harborRegistrySearchProviderFactory{})
}

var _ imagesearch.SearchProvider = &harborRegistrySearchProvider{}

type harborRegistrySearchProvider struct {
	HttpClient *http.Client `json:"-" yaml:"-"`
}

type searchResponse struct {
	Repository []repository `json:"repository"`
}

type repository struct {
	RepositoryName string `json:"repository_name"`
}

func (d harborRegistrySearchProvider) Search(imageName string, config imagesearch.SearchConfig) (*imagesearch.Results, error) {

	url := fmt.Sprintf("%s/%s", config.Host, fmt.Sprintf(harborSearchUrl, imageName))

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
		return nil, fmt.Errorf("search images failed with status code: %d, message: %s", resp.StatusCode, bytes)
	}

	searchResp := &searchResponse{}
	err = json.Unmarshal(bytes, searchResp)
	if err != nil {
		return nil, err
	}
	imageResult := &imagesearch.Results{
		Entries: make([]string, 0),
	}
	for _, v := range searchResp.Repository {
		imageResult.Entries = append(imageResult.Entries, v.RepositoryName)
	}

	imageResult.Total = int64(len(imageResult.Entries))

	return imageResult, nil
}

var _ imagesearch.SearchProviderFactory = &harborRegistrySearchProviderFactory{}

type harborRegistrySearchProviderFactory struct{}

func (d harborRegistrySearchProviderFactory) Type() string {
	return HarborRegisterProvider
}

func (d harborRegistrySearchProviderFactory) Create(_ map[string]interface{}) (imagesearch.SearchProvider, error) {
	var provider harborRegistrySearchProvider
	provider.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	return provider, nil
}

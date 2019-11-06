/*

 Copyright 2019 The KubeSphere Authors.

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
package prometheus

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/monitoring/v1alpha2"
	"net/http"
	"time"
)

type PrometheusClient struct {
	client            *http.Client
	endpoint          string
	secondaryEndpoint string
}

func NewPrometheusClient(options *PrometheusOptions) (*PrometheusClient, error) {
	return &PrometheusClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		endpoint:          options.Endpoint,
		secondaryEndpoint: options.SecondaryEndpoint,
	}, nil
}

func (c *PrometheusClient) QueryToK8SPrometheus(queryType string, params string) (apiResponse v1alpha2.APIResponse) {
	return c.query(c.endpoint, queryType, params)
}

func (c *PrometheusClient) QueryToK8SSystemPrometheus(queryType string, params string) (apiResponse v1alpha2.APIResponse) {
	return c.query(c.secondaryEndpoint, queryType, params)
}

var jsonIter = jsoniter.ConfigCompatibleWithStandardLibrary

func (c *PrometheusClient) query(endpoint string, queryType string, params string) (apiResponse v1alpha2.APIResponse) {
	url := fmt.Sprintf("%s/api/v1/%s?%s", endpoint, queryType, params)

	response, err := c.client.Get(url)
	if err != nil {
		klog.Error(err)
		apiResponse.Status = "error"
		return apiResponse
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		klog.Error(err)
		apiResponse.Status = "error"
		return apiResponse
	}

	err = jsonIter.Unmarshal(body, &apiResponse)
	if err != nil {
		klog.Errorf("fail to unmarshal prometheus query result: %s", err.Error())
		apiResponse.Status = "error"
		return apiResponse
	}

	return apiResponse
}

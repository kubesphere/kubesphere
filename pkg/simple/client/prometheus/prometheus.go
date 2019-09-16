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
	"io/ioutil"
	"k8s.io/klog"
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
			Timeout: time.Duration(3) * time.Second,
		},
		endpoint:          options.Endpoint,
		secondaryEndpoint: options.SecondaryEndpoint,
	}, nil
}

func (c *PrometheusClient) SendMonitoringRequest(queryType string, params string) string {
	return c.sendMonitoringRequest(c.endpoint, queryType, params)
}

func (c *PrometheusClient) SendSecondaryMonitoringRequest(queryType string, params string) string {
	return c.sendMonitoringRequest(c.secondaryEndpoint, queryType, params)
}

func (c *PrometheusClient) sendMonitoringRequest(endpoint string, queryType string, params string) string {
	epurl := endpoint + queryType + params
	response, err := c.client.Get(epurl)
	if err != nil {
		klog.Error(err)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			klog.Error(err)
		}
		return string(contents)
	}
	return ""
}

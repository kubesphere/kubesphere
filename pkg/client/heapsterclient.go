/*
Copyright 2018 The KubeSphere Authors.

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

package client

import (
	"net/http"

	"io/ioutil"
	"time"

	"github.com/antonholmquist/jason"
	"github.com/golang/glog"
)

const (
	DefaultHeapsterScheme  = "http"
	DefaultHeapsterService = "heapster.kube-system.svc.cluster.local" //"heapster"
	DefaultHeapsterPort    = "80"                                     // use the first exposed port on the service
	HeapsterApiPath        = "/api/v1/model"
	HeapsterEndpointUrl    = DefaultHeapsterScheme + "://" + DefaultHeapsterService + ":" + DefaultHeapsterPort + HeapsterApiPath
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

// Get heapster response in python-like dictionary
func GetHeapsterMetricsJson(url string) *jason.Object {

	response, err := httpClient.Get(HeapsterEndpointUrl + url)

	var data *jason.Object
	if err != nil {
		glog.Error(url, err)
	} else {
		defer response.Body.Close()

		data, err = jason.NewObjectFromReader(response.Body)

		if err != nil {
			glog.Error(url, err)
		}
	}

	// return empty json in case of error response from es-node
	if data == nil {
		emptyJSON := `{}`
		data, _ = jason.NewObjectFromBytes([]byte(emptyJSON))
	}

	return data
}

func GetHeapsterMetrics(url string) string {
	response, err := httpClient.Get(HeapsterEndpointUrl + url)
	if err != nil {
		glog.Error(err)

	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			glog.Error(err)
		}

		return string(contents)
	}
	return ""
}
func GetCAdvisorMetrics(nodeAddr string) string {

	response, err := httpClient.Get("http://" + nodeAddr + ":10255/stats/summary")
	if err != nil {
		glog.Error(err)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			glog.Error(err)
		}

		return string(contents)
	}
	return ""
}

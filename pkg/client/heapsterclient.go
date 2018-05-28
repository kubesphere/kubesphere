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

	"github.com/golang/glog"
	"io/ioutil"
	"os"
)

const (
	DefaultHeapsterScheme  = "http"
	DefaultHeapsterService = "heapster" //"heapster"
	DefaultHeapsterPort    = "80"        // use the first exposed port on the service
)

var (
	prefix = "/api/v1/model"
)

func GetHeapsterMetrics(url string) string {
	//glog.Info("Querying data from " + DefaultHeapsterScheme + "://" + DefaultHeapsterService + ":" + DefaultHeapsterPort + prefix + url)
	response, err := http.Get(DefaultHeapsterScheme + "://" + DefaultHeapsterService + ":" + DefaultHeapsterPort + prefix + url)
	if err != nil {
		glog.Error(err)
		os.Exit(1)
	} else {
		defer response.Body.Close()

		contents, err := ioutil.ReadAll(response.Body)

		if err != nil {
			glog.Error(err)
			os.Exit(1)
		}

		return string(contents)
	}
	return ""
}

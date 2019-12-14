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

package tracing

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"io/ioutil"
	"log"
	"net/http"
)

var JaegerQueryUrl = "http://jaeger-query.istio-system.svc:16686/jaeger"

func GetServiceTracing(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	service := request.PathParameter("service")

	serviceName := fmt.Sprintf("%s.%s", service, namespace)

	url := fmt.Sprintf("%s/api/traces?%s&service=%s", JaegerQueryUrl, request.Request.URL.RawQuery, serviceName)

	resp, err := http.Get(url)

	if err != nil {
		log.Printf("query jaeger faile with err %v", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if err != nil {
		log.Printf("read response error : %v", err)
		response.WriteError(http.StatusInternalServerError, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		log.Printf("write response failed %v", err)
	}
}

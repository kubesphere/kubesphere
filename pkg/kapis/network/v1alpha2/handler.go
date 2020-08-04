/*
Copyright 2020 KubeSphere Authors

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

package v1alpha2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/emicklei/go-restful"
	"k8s.io/klog"

	"kubesphere.io/kubesphere/pkg/api"
)

const ScopeQueryUrl = "http://%s/api/topology/services"

type handler struct {
	weaveScopeHost string
}

func (h *handler) getScopeUrl() string {
	return fmt.Sprintf(ScopeQueryUrl, h.weaveScopeHost)
}

func (h *handler) getNamespaceTopology(request *restful.Request, response *restful.Response) {
	var query = url.Values{
		"namespace": []string{request.PathParameter("namespace")},
		"timestamp": request.QueryParameters("timestamp"),
	}
	var u = fmt.Sprintf("%s?%s", h.getScopeUrl(), query.Encode())

	resp, err := http.Get(u)

	if err != nil {
		klog.Errorf("query scope faile with err %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		klog.Errorf("read response error : %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		klog.Errorf("write response failed %v", err)
	}
}

func (h *handler) getNamespaceNodeTopology(request *restful.Request, response *restful.Response) {
	var query = url.Values{
		"namespace": []string{request.PathParameter("namespace")},
		"timestamp": request.QueryParameters("timestamp"),
	}
	var u = fmt.Sprintf("%s/%s?%s", h.getScopeUrl(), request.PathParameter("node_id"), query.Encode())

	resp, err := http.Get(u)

	if err != nil {
		klog.Errorf("query scope faile with err %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		klog.Errorf("read response error : %v", err)
		api.HandleInternalError(response, nil, err)
		return
	}

	// need to set header for proper response
	response.Header().Set("Content-Type", "application/json")
	_, err = response.Write(body)

	if err != nil {
		klog.Errorf("write response failed %v", err)
	}
}

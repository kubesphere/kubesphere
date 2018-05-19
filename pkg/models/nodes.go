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

package models

import (
	"github.com/emicklei/go-restful"

	"net/http"

	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/client"
)

type ResultMessage struct {
	Ret  int         `json:"ret"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func HandleNodes(request *restful.Request, response *restful.Response) {

	var result ResultMessage


	data := make(map[string]string)


	data["output"] = client.GetHeapsterMetrics("http://139.198.0.79/api/monitor/v1/model/namespaces/kube-system/pods/qingcloud-volume-provisioner-i-o5pmakm7/metrics/cpu/request")

	result.Data = data
	result.Ret = http.StatusOK
	result.Msg = "success"
	glog.Infoln(result)
	response.WriteAsJson(result)
}

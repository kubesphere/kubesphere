// Copyright 2018 The Kubesphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kubeconfig

import (
	"net/http"

	"github.com/emicklei/go-restful"

	"kubesphere.io/kubesphere/pkg/models"
)

const DefaultServiceAccount = "default"

type Config struct {
	Certificate string
	Server      string
	User        string
	Token       string
}

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath).To(handleKubeconfig))

}

func handleKubeconfig(req *restful.Request, resp *restful.Response) {

	ns := req.PathParameter("namespace")

	kubectlConfig, err := models.GetKubeConfig(ns)

	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}

	resp.WriteEntity(kubectlConfig)
}

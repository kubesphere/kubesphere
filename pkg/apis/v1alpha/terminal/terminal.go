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

package terminal

import (
	"github.com/emicklei/go-restful"

	"net/http"

	"kubesphere.io/kubesphere/pkg/models"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath).To(handleExecShell))

}

func handleExecShell(req *restful.Request, resp *restful.Response) {
	res, err := models.HandleExecShell(req)
	if err != nil {
		resp.WriteError(http.StatusInternalServerError, err)
	}

	resp.WriteEntity(res)

}

func RegisterWebSocketHandler(container *restful.Container, path string) {
	handler := models.CreateTerminalHandler(path[0 : len(path)-1])
	container.Handle(path, handler)
}

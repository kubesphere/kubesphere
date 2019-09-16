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
package terminal

import (
	"github.com/emicklei/go-restful"
	"gopkg.in/igm/sockjs-go.v2/sockjs"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"net/http"
)

// TerminalResponse is sent by handleExecShell. The Id is a random session id that binds the original REST request and the SockJS connection.
// Any clientapi in possession of this Id can hijack the terminal session.
type TerminalResponse struct {
	Id string `json:"id"`
}

// CreateAttachHandler is called from main for /api/sockjs
func NewTerminalHandler(path string) http.Handler {
	return sockjs.NewHandler(path, sockjs.DefaultOptions, terminal.HandleTerminalSession)
}

// Handles execute shell API call
func CreateTerminalSession(request *restful.Request, resp *restful.Response) {

	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	shell := request.QueryParameter("shell")

	sessionId, err := terminal.NewSession(shell, namespace, podName, containerName)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	TerminalResponse := &TerminalResponse{Id: sessionId}
	resp.WriteAsJson(TerminalResponse)
}

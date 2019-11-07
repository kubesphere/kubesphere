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
	"github.com/gorilla/websocket"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/terminal"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Handles execute shell API call
func HandleTerminalSession(request *restful.Request, resp *restful.Response) {

	namespace := request.PathParameter("namespace")
	podName := request.PathParameter("pod")
	containerName := request.QueryParameter("container")
	shell := request.QueryParameter("shell")

	conn, err := upgrader.Upgrade(resp.ResponseWriter, request.Request, nil)
	if err != nil {
		klog.Warning(err)
		return
	}

	terminal.HandleSession(shell, namespace, podName, containerName, conn)
}

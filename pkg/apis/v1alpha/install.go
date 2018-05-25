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

package v1alpha

import (
        "github.com/emicklei/go-restful"
        "kubesphere.io/kubesphere/pkg/apis/v1alpha/nodes"
        "kubesphere.io/kubesphere/pkg/apis/v1alpha/kubeconfig"
        "kubesphere.io/kubesphere/pkg/apis/v1alpha/kubectl"
		"kubesphere.io/kubesphere/pkg/apis/v1alpha/terminal"
)

func init() {
	ws := new(restful.WebService)
    ws.Path("/api/v1alpha")

    nodes.Register(ws,"/nodes")
    kubeconfig.Register(ws, "/namespaces/{namespace}/kubeconfig")
    kubectl.Register(ws, "/namespaces/{namespace}/kubectl")
    terminal.Register(ws, "/namespaces/{namespace}/pod/{pod}/shell/{container}")

    // add webservice to default container
    restful.Add(ws)

    // add websocket handler to default container
    terminal.RegisterWebSocketHandler(restful.DefaultContainer, "/api/v1alpha/sockjs/")

}


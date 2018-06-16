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

	"kubesphere.io/kubesphere/pkg/apis/v1alpha/components"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/containers"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/iam"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/nodes"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/pods"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/quota"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/registries"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/resources"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/routes"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/storage"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/terminal"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/users"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/volumes"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/workloadstatus"
)

func init() {

	ws := new(restful.WebService)
	ws.Path("/api/v1alpha1")

	registries.Register(ws, "/registries")
	storage.Register(ws, "/storage")
	volumes.Register(ws, "/volumes")
	nodes.Register(ws, "/nodes")
	pods.Register(ws)
	containers.Register(ws)
	iam.Register(ws)
	components.Register(ws, "/components")
	routes.Register(ws)
	user.Register(ws, "/users/{user}")
	terminal.Register(ws, "/namespaces/{namespace}/pod/{pod}/shell/{container}")
	workloadstatus.Register(ws, "/status")
	quota.Register(ws, "/quota")
	resources.Register(ws, "/resources")

	// add webservice to default container
	restful.Add(ws)

	// add websocket handler to default container
	terminal.RegisterWebSocketHandler(restful.DefaultContainer, "/api/v1alpha1/sockjs/")
}

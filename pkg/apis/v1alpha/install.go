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
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/containers"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/kubeconfig"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/kubectl"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/nodes"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/pods"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/registries"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/storage"
	"kubesphere.io/kubesphere/pkg/apis/v1alpha/volumes"
)

func init() {

	ws := new(restful.WebService)
	ws.Path("/api/v1alpha1")

	kubeconfig.Register(ws, "/namespaces/{namespace}/kubeconfig")
	kubectl.Register(ws, "/namespaces/{namespace}/kubectl")
	registries.Register(ws, "/registries")
	storage.Register(ws, "/storage")
	volumes.Register(ws, "/volumes")
	nodes.Register(ws, "/nodes")
	pods.Register(ws)
	containers.Register(ws)
	// add webservice to default container
	restful.Add(ws)

}

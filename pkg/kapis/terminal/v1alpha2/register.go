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

package v1alpha2

import (
	"github.com/emicklei/go-restful"
	restfulspec "github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"kubesphere.io/kubesphere/pkg/apiserver/authorization/authorizer"

	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/terminal"
)

const (
	GroupName = "terminal.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, client kubernetes.Interface, authorizer authorizer.Authorizer, config *rest.Config, options *terminal.Options) error {

	webservice := runtime.NewWebService(GroupVersion)

	handler := newTerminalHandler(client, authorizer, config, options)

	webservice.Route(webservice.GET("/namespaces/{namespace}/pods/{pod}/exec").
		To(handler.handleTerminalSession).
		Param(webservice.PathParameter("namespace", "namespace of which the pod located in")).
		Param(webservice.PathParameter("pod", "name of the pod")).
		Doc("create terminal session").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TerminalTag}).
		Writes(models.PodInfo{}))

	//Add new Route to support shell access to the node
	webservice.Route(webservice.GET("/nodes/{nodename}/exec").
		To(handler.handleShellAccessToNode).
		Param(webservice.PathParameter("nodename", "name of cluster node")).
		Doc("create shell access to node session").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TerminalTag}).
		Writes(models.PodInfo{}))

	c.Add(webservice)

	return nil
}

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
	"github.com/emicklei/go-restful-openapi"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
)

const (
	GroupName = "terminal.kubesphere.io"
)

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

func AddToContainer(c *restful.Container, client kubernetes.Interface, config *rest.Config) error {

	webservice := runtime.NewWebService(GroupVersion)

	handler := newTerminalHandler(client, config)

	webservice.Route(webservice.GET("/namespaces/{namespace}/pods/{pod}").
		To(handler.handleTerminalSession).
		Param(webservice.PathParameter("namespace", "namespace of which the pod located in")).
		Param(webservice.PathParameter("pod", "name of the pod")).
		Doc("create terminal session").
		Metadata(restfulspec.KeyOpenAPITags, []string{constants.TerminalTag}).
		Writes(models.PodInfo{}))

	c.Add(webservice)

	return nil
}

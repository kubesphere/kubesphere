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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/kubesphere/pkg/apiserver/operations"
	"kubesphere.io/kubesphere/pkg/apiserver/runtime"
	"kubesphere.io/kubesphere/pkg/errors"
	"net/http"
)

const GroupName = "operations.kubesphere.io"

var GroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1alpha2"}

var (
	WebServiceBuilder = runtime.NewContainerBuilder(addWebService)
	AddToContainer    = WebServiceBuilder.AddToContainer
)

func addWebService(c *restful.Container) error {

	ok := "ok"
	webservice := runtime.NewWebService(GroupVersion)

	webservice.Route(webservice.POST("/nodes/{node}/drainage").
		To(operations.DrainNode).
		Deprecate().
		Doc("remove a node from service, safely evict all of your pods from a node and you can power down the node. More info: https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/").
		Param(webservice.PathParameter("node", "node name")).
		Returns(http.StatusOK, ok, errors.Error{}))

	webservice.Route(webservice.POST("/namespaces/{namespace}/jobs/{job}").
		To(operations.RerunJob).
		Doc("Rerun job whether the job is complete or not").
		Deprecate().
		Param(webservice.PathParameter("job", "job name")).
		Param(webservice.PathParameter("namespace", "the name of the namespace where the job runs in")).
		Param(webservice.QueryParameter("action", "action must be \"rerun\"")).
		Returns(http.StatusOK, ok, errors.Error{}))

	c.Add(webservice)

	return nil
}

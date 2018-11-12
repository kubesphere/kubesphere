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
package logging

import (
	//"strings"

	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful-openapi"

	//"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/models/log"
)

func (u LoggingResource) loggingQuery(request *restful.Request, response *restful.Response) {
	res := log.LogQuery(request)
	response.WriteAsJson(res)
}

type LoggingResource struct {
}

func Register(ws *restful.WebService, subPath string) {
	tags := []string{"logging apis"}
	u := LoggingResource{}

	ws.Route(ws.GET(subPath+"/clusters").To(u.loggingQuery).
		Filter(route.RouteLogging).
		Doc("cluster level log query").
		Param(ws.QueryParameter("log_query", "log query keywords").DataType("string").Required(false)).
		Param(ws.QueryParameter("start", "range start time").DataType("string").Required(false)).
		Param(ws.QueryParameter("end", "range end time").DataType("string").Required(false)).
		Metadata(restfulspec.KeyOpenAPITags, tags)).
		Consumes(restful.MIME_JSON, restful.MIME_XML).
		Produces(restful.MIME_JSON)
}

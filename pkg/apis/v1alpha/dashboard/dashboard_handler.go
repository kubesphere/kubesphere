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

package dashboard

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/filter/route"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models"
	"net/http"
	"github.com/golang/glog"
)

func Register(ws *restful.WebService, subPath string) {

	ws.Route(ws.GET(subPath + "/pods/{query}").To(handlerPodsData).Filter(route.RouteLogging)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

}

func handlerPodsData(request *restful.Request, response *restful.Response)  {

	var result constants.ResultMessage

	querytype := request.PathParameter("query")

	data,err := models.QueryPodsData(querytype)

	if err != nil{

		glog.Error(err)

		response.WriteError(http.StatusInternalServerError,err)

	}

	result.Data = data

	response.WriteAsJson(result)

}
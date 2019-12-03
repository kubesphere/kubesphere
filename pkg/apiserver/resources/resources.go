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
package resources

import (
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"net/http"

	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
)

func ListNamespacedResources(req *restful.Request, resp *restful.Response) {
	ListResources(req, resp)
}

func ListResources(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	resourceName := req.PathParameter("resources")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := params.GetStringValueWithDefault(req, params.OrderByParam, resources.CreateTime)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := resources.ListResources(namespace, resourceName, conditions, orderBy, reverse, limit, offset)

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteAsJson(result)
}

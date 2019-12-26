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

package serverless

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models/resources"
	"kubesphere.io/kubesphere/pkg/models/serverless"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"net/http"
)

func ListServices(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	fmt.Println("listing services in", namespace)
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = resources.CreateTime
		reverse = true
	}

	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := serverless.ListServices(namespace, conditions, orderBy, reverse, limit, offset)
	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	_ = resp.WriteAsJson(result)
}

func ListConfigurations(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	fmt.Println("listing configurations in", namespace)
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = resources.CreateTime
		reverse = true
	}

	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := serverless.ListConfigurations(namespace, conditions, orderBy, reverse, limit, offset)
	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	_ = resp.WriteAsJson(result)
}

func ListRevisions(req *restful.Request, resp *restful.Response) {
	namespace := req.PathParameter("namespace")
	fmt.Println("listing revisions in", namespace)
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)

	if orderBy == "" {
		orderBy = resources.CreateTime
		reverse = true
	}

	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := serverless.ListRevisions(namespace, conditions, orderBy, reverse, limit, offset)
	if err != nil {
		_ = resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	_ = resp.WriteAsJson(result)
}

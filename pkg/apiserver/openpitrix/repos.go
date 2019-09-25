/*
 *
 * Copyright 2019 The KubeSphere Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 * /
 */

package openpitrix

import (
	"github.com/emicklei/go-restful"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/models/openpitrix"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/server/params"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
	"strconv"
)

func CreateRepo(req *restful.Request, resp *restful.Response) {
	createRepoRequest := &openpitrix.CreateRepoRequest{}
	err := req.ReadEntity(createRepoRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	validate, _ := strconv.ParseBool(req.QueryParameter("validate"))

	var result interface{}

	if validate {
		validateRepoRequest := &openpitrix.ValidateRepoRequest{
			Type:       createRepoRequest.Type,
			Url:        createRepoRequest.URL,
			Credential: createRepoRequest.Credential,
		}
		result, err = openpitrix.ValidateRepo(validateRepoRequest)
	} else {
		result, err = openpitrix.CreateRepo(createRepoRequest)
	}

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func DoRepoAction(req *restful.Request, resp *restful.Response) {
	repoActionRequest := &openpitrix.RepoActionRequest{}
	repoId := req.PathParameter("repo")
	err := req.ReadEntity(repoActionRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = openpitrix.DoRepoAction(repoId, repoActionRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DeleteRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	err := openpitrix.DeleteRepo(repoId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func ModifyRepo(req *restful.Request, resp *restful.Response) {
	var updateRepoRequest openpitrix.ModifyRepoRequest
	repoId := req.PathParameter("repo")
	err := req.ReadEntity(&updateRepoRequest)
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	err = openpitrix.PatchRepo(repoId, &updateRepoRequest)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.InvalidArgument {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(errors.None)
}

func DescribeRepo(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")

	result, err := openpitrix.DescribeRepo(repoId)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}
	if status.Code(err) == codes.NotFound {
		resp.WriteHeaderAndEntity(http.StatusNotFound, errors.Wrap(err))
		return
	}
	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}
func ListRepos(req *restful.Request, resp *restful.Response) {
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	orderBy := req.QueryParameter(params.OrderByParam)
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))
	reverse := params.ParseReverse(req)
	if orderBy == "" {
		orderBy = "create_time"
		reverse = true
	}

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := openpitrix.ListRepos(conditions, orderBy, reverse, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

func ListRepoEvents(req *restful.Request, resp *restful.Response) {
	repoId := req.PathParameter("repo")
	conditions, err := params.ParseConditions(req.QueryParameter(params.ConditionsParam))
	limit, offset := params.ParsePaging(req.QueryParameter(params.PagingParam))

	if err != nil {
		resp.WriteHeaderAndEntity(http.StatusBadRequest, errors.Wrap(err))
		return
	}

	result, err := openpitrix.ListRepoEvents(repoId, conditions, limit, offset)

	if _, notEnabled := err.(client.ClientSetNotEnabledError); notEnabled {
		resp.WriteHeaderAndEntity(http.StatusNotImplemented, errors.Wrap(err))
		return
	}

	if err != nil {
		klog.Errorln(err)
		resp.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	resp.WriteEntity(result)
}

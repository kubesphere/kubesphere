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

package git

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/models/git"
	"kubesphere.io/kubesphere/pkg/server/errors"
)

func GitReadVerify(request *restful.Request, response *restful.Response) {

	authInfo := git.AuthInfo{}

	err := request.ReadEntity(&authInfo)
	ns := request.PathParameter("namespace")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = git.GitReadVerify(ns, authInfo)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(errors.None)
}

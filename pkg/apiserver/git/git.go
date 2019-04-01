package git

import (
	"net/http"

	"github.com/emicklei/go-restful"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/git"
)

func GitReadVerify(request *restful.Request, response *restful.Response) {

	authInfo := git.AuthInfo{}

	err := request.ReadEntity(&authInfo)
	ns := request.PathParameter("namespace")
	name := request.PathParameter("name")
	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	err = git.GitReadVerify(ns, name, authInfo)

	if err != nil {
		response.WriteHeaderAndEntity(http.StatusInternalServerError, errors.Wrap(err))
		return
	}

	response.WriteAsJson(errors.None)
}

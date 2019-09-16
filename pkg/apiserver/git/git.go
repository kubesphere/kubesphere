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

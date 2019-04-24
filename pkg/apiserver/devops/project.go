package devops

import (
	"github.com/emicklei/go-restful"
	"github.com/golang/glog"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/errors"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"net/http"
)

func GetDevOpsProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)

	err := devops.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err := devops.GetProject(projectId)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func UpdateProjectHandler(request *restful.Request, resp *restful.Response) {

	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	var project *devops.DevOpsProject
	err := request.ReadEntity(&project)
	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusBadRequest, err.Error()), resp)
		return
	}
	project.ProjectId = projectId
	err = devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	project, err = devops.UpdateProject(project)

	if err != nil {
		glog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}

	resp.WriteAsJson(project)
	return
}

func GetDevOpsProjectDefaultRoles(request *restful.Request, resp *restful.Response) {
	resp.WriteAsJson(devops.DefaultRoles)
	return
}

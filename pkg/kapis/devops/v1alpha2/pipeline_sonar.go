package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

func (h PipelineSonarHandler) GetPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	sonarStatus, err := h.pipelineSonarGetter.GetPipelineSonar(projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

func (h PipelineSonarHandler) GetMultiBranchesPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	branchId := request.PathParameter("branch")
	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(restful.NewError(http.StatusForbidden, err.Error()), resp)
		return
	}
	sonarStatus, err := h.pipelineSonarGetter.GetMultiBranchPipelineSonar(projectId, pipelineId, branchId)
	if err != nil {
		klog.Errorf("%+v", err)
		errors.ParseSvcErr(err, resp)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

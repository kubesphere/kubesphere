package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

func (h PipelineSonarHandler) GetPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	username := request.HeaderParameter(constants.UserNameHeader)
	pipelineId := request.PathParameter("pipeline")
	err := h.projectOperator.CheckProjectUserInRole(username, projectId, devops.AllRoleSlice)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleForbidden(resp, err)
		return
	}
	sonarStatus, err := h.pipelineSonarGetter.GetPipelineSonar(projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, err)
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
		api.HandleForbidden(resp, err)
		return
	}
	sonarStatus, err := h.pipelineSonarGetter.GetMultiBranchPipelineSonar(projectId, pipelineId, branchId)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, err)
		return
	}
	resp.WriteAsJson(sonarStatus)
}
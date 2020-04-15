package v1alpha2

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
)

func (h PipelineSonarHandler) GetPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	pipelineId := request.PathParameter("pipeline")
	sonarStatus, err := h.pipelineSonarGetter.GetPipelineSonar(projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

func (h PipelineSonarHandler) GetMultiBranchesPipelineSonarStatusHandler(request *restful.Request, resp *restful.Response) {
	projectId := request.PathParameter("devops")
	pipelineId := request.PathParameter("pipeline")
	branchId := request.PathParameter("branch")
	sonarStatus, err := h.pipelineSonarGetter.GetMultiBranchPipelineSonar(projectId, pipelineId, branchId)
	if err != nil {
		klog.Errorf("%+v", err)
		api.HandleInternalError(resp, nil, err)
		return
	}
	resp.WriteAsJson(sonarStatus)
}

/*
Copyright 2020 The KubeSphere Authors.

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

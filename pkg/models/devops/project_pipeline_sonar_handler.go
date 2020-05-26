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

package devops

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/server/errors"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/sonarqube"
	"net/http"
)

type PipelineSonarGetter interface {
	GetPipelineSonar(projectId, pipelineId string) ([]*sonarqube.SonarStatus, error)
	GetMultiBranchPipelineSonar(projectId, pipelineId, branchId string) ([]*sonarqube.SonarStatus, error)
}
type pipelineSonarGetter struct {
	devops.BuildGetter
	sonarqube.SonarInterface
}

func NewPipelineSonarGetter(devopClient devops.BuildGetter, sonarClient sonarqube.SonarInterface) PipelineSonarGetter {
	return &pipelineSonarGetter{
		BuildGetter:    devopClient,
		SonarInterface: sonarClient,
	}
}

func (g *pipelineSonarGetter) GetPipelineSonar(projectId, pipelineId string) ([]*sonarqube.SonarStatus, error) {

	build, err := g.GetProjectPipelineBuildByType(projectId, pipelineId, devops.LastBuild)
	if err != nil && errors.GetServiceErrorCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, err
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}
	var taskIds []string
	for _, action := range build.Actions {
		if action.ClassName == sonarqube.SonarAnalysisActionClass {
			taskIds = append(taskIds, action.SonarTaskId)
		}
	}
	var sonarStatus []*sonarqube.SonarStatus

	if len(taskIds) != 0 {
		sonarStatus, err = g.GetSonarResultsByTaskIds(taskIds...)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	} else if len(taskIds) == 0 {
		build, err := g.GetProjectPipelineBuildByType(projectId, pipelineId, devops.LastCompletedBuild)
		if err != nil && errors.GetServiceErrorCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(errors.GetServiceErrorCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		for _, action := range build.Actions {
			if action.ClassName == sonarqube.SonarAnalysisActionClass {
				taskIds = append(taskIds, action.SonarTaskId)
			}
		}
		sonarStatus, err = g.GetSonarResultsByTaskIds(taskIds...)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}

	}
	return sonarStatus, nil
}

func (g *pipelineSonarGetter) GetMultiBranchPipelineSonar(projectId, pipelineId, branchId string) ([]*sonarqube.SonarStatus, error) {

	build, err := g.GetMultiBranchPipelineBuildByType(projectId, pipelineId, branchId, devops.LastBuild)
	if err != nil && errors.GetServiceErrorCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(errors.GetServiceErrorCode(err), err.Error())
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}
	var taskIds []string
	for _, action := range build.Actions {
		if action.ClassName == sonarqube.SonarAnalysisActionClass {
			taskIds = append(taskIds, action.SonarTaskId)
		}
	}
	var sonarStatus []*sonarqube.SonarStatus

	if len(taskIds) != 0 {
		sonarStatus, err = g.GetSonarResultsByTaskIds(taskIds...)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	} else if len(taskIds) == 0 {
		build, err := g.GetMultiBranchPipelineBuildByType(projectId, pipelineId, branchId, devops.LastCompletedBuild)
		if err != nil && errors.GetServiceErrorCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(errors.GetServiceErrorCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		for _, action := range build.Actions {
			if action.ClassName == sonarqube.SonarAnalysisActionClass {
				taskIds = append(taskIds, action.SonarTaskId)
			}
		}
		sonarStatus, err = g.GetSonarResultsByTaskIds(taskIds...)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}

	}
	return sonarStatus, nil
}

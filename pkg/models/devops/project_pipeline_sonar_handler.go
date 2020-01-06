package devops

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	cs "kubesphere.io/kubesphere/pkg/simple/client"
	"net/http"
)

type PipelineSonarGetter interface {
	GetPipelineSonar(projectId, pipelineId string) ([]*SonarStatus, error)
	GetMultiBranchPipelineSonar(projectId, pipelineId, branchId string) ([]*SonarStatus, error)
}
type pipelineSonarGetter struct {
}

func GetPipelineSonar(projectId, pipelineId string) ([]*SonarStatus, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	job, err := jenkinsClient.GetJob(pipelineId, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	build, err := job.GetLastBuild()
	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}

	sonarStatus, err := getBuildSonarResults(build)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
	if len(sonarStatus) == 0 {
		build, err := job.GetLastCompletedBuild()
		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		sonarStatus, err = getBuildSonarResults(build)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	}
	return sonarStatus, nil
}

func GetMultiBranchPipelineSonar(projectId, pipelineId, branchId string) ([]*SonarStatus, error) {
	devops, err := cs.ClientSets().Devops()
	if err != nil {
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	jenkinsClient := devops.Jenkins()

	job, err := jenkinsClient.GetJob(branchId, projectId, pipelineId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	build, err := job.GetLastBuild()
	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	} else if err != nil {
		klog.Errorf("%+v", err)
		return nil, nil
	}

	sonarStatus, err := getBuildSonarResults(build)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}
	if len(sonarStatus) == 0 {
		build, err := job.GetLastCompletedBuild()
		if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		} else if err != nil {
			klog.Errorf("%+v", err)
			return nil, nil
		}
		sonarStatus, err = getBuildSonarResults(build)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, restful.NewError(http.StatusBadRequest, err.Error())
		}
	}
	return sonarStatus, nil
}

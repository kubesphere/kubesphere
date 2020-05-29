package jenkins

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
)

type DevOpsProjectRoleResponse struct {
	ProjectRole *ProjectRole
	Err         error
}

func (j *Jenkins) CreateDevOpsProject(projectId string) (string, error) {
	_, err := j.CreateFolder(projectId, "")
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
	}
	return projectId, nil
}

func (j *Jenkins) DeleteDevOpsProject(projectId string) error {
	_, err := j.DeleteJob(projectId)

	if err != nil && devops.GetDevOpsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())
	}
	return nil
}

func (j *Jenkins) GetDevOpsProject(projectId string) (string, error) {
	job, err := j.GetJob(projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(devops.GetDevOpsStatusCode(err), err.Error())

	}
	return job.GetName(), nil
}

package jenkins

import (
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"net/http"
	"sync"
)

type DevOpsProjectRoleResponse struct {
	ProjectRole *ProjectRole
	Err         error
}

func (j *Jenkins) CreateDevOpsProject(username string, project *v1alpha2.DevOpsProject) (*v1alpha2.DevOpsProject, error) {
	_, err := j.CreateFolder(project.ProjectId, project.Description)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	var addRoleCh = make(chan *DevOpsProjectRoleResponse, 8)
	var addRoleWg sync.WaitGroup
	for role, permission := range JenkinsProjectPermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission ProjectPermissionIds) {
			_, err := j.AddProjectRole(GetProjectRoleName(project.ProjectId, role),
				GetProjectRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	for role, permission := range JenkinsPipelinePermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission ProjectPermissionIds) {
			_, err := j.AddProjectRole(GetPipelineRoleName(project.ProjectId, role),
				GetPipelineRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	addRoleWg.Wait()
	close(addRoleCh)
	for addRoleResponse := range addRoleCh {
		if addRoleResponse.Err != nil {
			klog.Errorf("%+v", addRoleResponse.Err)
			return nil, restful.NewError(GetJenkinsStatusCode(addRoleResponse.Err), addRoleResponse.Err.Error())
		}
	}

	globalRole, err := j.GetGlobalRole(JenkinsAllUserRoleName)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	if globalRole == nil {
		_, err := j.AddGlobalRole(JenkinsAllUserRoleName, GlobalPermissionIds{
			GlobalRead: true,
		}, true)
		if err != nil {
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
	}
	err = globalRole.AssignRole(username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	projectRole, err := j.GetProjectRole(GetProjectRoleName(project.ProjectId, devops.ProjectOwner))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = projectRole.AssignRole(username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	pipelineRole, err := j.GetProjectRole(GetPipelineRoleName(project.ProjectId, devops.ProjectOwner))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = pipelineRole.AssignRole(username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return project, nil
}

func (j *Jenkins) DeleteDevOpsProject(projectId string) error {
	_, err := j.DeleteJob(projectId)

	if err != nil && GetJenkinsStatusCode(err) != http.StatusNotFound {
		klog.Errorf("%+v", err)
		return restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	roleNames := make([]string, 0)
	for role := range JenkinsProjectPermissionMap {
		roleNames = append(roleNames, GetProjectRoleName(projectId, role))
		roleNames = append(roleNames, GetPipelineRoleName(projectId, role))
	}
	err = j.DeleteProjectRoles(roleNames...)
	if err != nil {
		klog.Errorf("%+v", err)
		return restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return nil
}

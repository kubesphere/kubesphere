package jenkins

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
)

func GetProjectRoleName(projectId, role string) string {
	return fmt.Sprintf("%s-%s-project", projectId, role)
}

func GetPipelineRoleName(projectId, role string) string {
	return fmt.Sprintf("%s-%s-pipeline", projectId, role)
}

func GetProjectRolePattern(projectId string) string {
	return fmt.Sprintf("^%s$", projectId)
}

func GetPipelineRolePattern(projectId string) string {
	return fmt.Sprintf("^%s/.*", projectId)
}

var JenkinsOwnerProjectPermissionIds = &ProjectPermissionIds{
	CredentialCreate:        true,
	CredentialDelete:        true,
	CredentialManageDomains: true,
	CredentialUpdate:        true,
	CredentialView:          true,
	ItemBuild:               true,
	ItemCancel:              true,
	ItemConfigure:           true,
	ItemCreate:              true,
	ItemDelete:              true,
	ItemDiscover:            true,
	ItemMove:                true,
	ItemRead:                true,
	ItemWorkspace:           true,
	RunDelete:               true,
	RunReplay:               true,
	RunUpdate:               true,
	SCMTag:                  true,
}

var JenkinsProjectPermissionMap = map[string]ProjectPermissionIds{
	devops.ProjectOwner: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	devops.ProjectMaintainer: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              true,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	devops.ProjectDeveloper: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  false,
	},
	devops.ProjectReporter: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               false,
		ItemCancel:              false,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           false,
		RunDelete:               false,
		RunReplay:               false,
		RunUpdate:               false,
		SCMTag:                  false,
	},
}

var JenkinsPipelinePermissionMap = map[string]ProjectPermissionIds{
	devops.ProjectOwner: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	devops.ProjectMaintainer: {
		CredentialCreate:        true,
		CredentialDelete:        true,
		CredentialManageDomains: true,
		CredentialUpdate:        true,
		CredentialView:          true,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           true,
		ItemCreate:              true,
		ItemDelete:              true,
		ItemDiscover:            true,
		ItemMove:                true,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  true,
	},
	devops.ProjectDeveloper: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               true,
		ItemCancel:              true,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           true,
		RunDelete:               true,
		RunReplay:               true,
		RunUpdate:               true,
		SCMTag:                  false,
	},
	devops.ProjectReporter: {
		CredentialCreate:        false,
		CredentialDelete:        false,
		CredentialManageDomains: false,
		CredentialUpdate:        false,
		CredentialView:          false,
		ItemBuild:               false,
		ItemCancel:              false,
		ItemConfigure:           false,
		ItemCreate:              false,
		ItemDelete:              false,
		ItemDiscover:            true,
		ItemMove:                false,
		ItemRead:                true,
		ItemWorkspace:           false,
		RunDelete:               false,
		RunReplay:               false,
		RunUpdate:               false,
		SCMTag:                  false,
	},
}

func (j *Jenkins) AddProjectMember(membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error) {
	globalRole, err := j.GetGlobalRole(JenkinsAllUserRoleName)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	if globalRole == nil {
		_, err := j.AddGlobalRole(JenkinsAllUserRoleName, GlobalPermissionIds{
			GlobalRead: true,
		}, true)
		if err != nil {
			klog.Errorf("failed to create jenkins global role %+v", err)
			return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
		}
	}
	err = globalRole.AssignRole(membership.Username)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	projectRole, err := j.GetProjectRole(GetProjectRoleName(membership.ProjectId, membership.Role))
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = projectRole.AssignRole(membership.Username)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	pipelineRole, err := j.GetProjectRole(GetPipelineRoleName(membership.ProjectId, membership.Role))
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = pipelineRole.AssignRole(membership.Username)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return membership, nil
}

func (j *Jenkins) UpdateProjectMember(oldMembership, newMembership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error) {
	oldProjectRole, err := j.GetProjectRole(GetProjectRoleName(oldMembership.ProjectId, oldMembership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	err = oldProjectRole.UnAssignRole(newMembership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	oldPipelineRole, err := j.GetProjectRole(GetPipelineRoleName(oldMembership.ProjectId, oldMembership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	err = oldPipelineRole.UnAssignRole(newMembership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	projectRole, err := j.GetProjectRole(GetProjectRoleName(oldMembership.ProjectId, newMembership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	err = projectRole.AssignRole(newMembership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	pipelineRole, err := j.GetProjectRole(GetPipelineRoleName(oldMembership.ProjectId, newMembership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	err = pipelineRole.AssignRole(newMembership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return newMembership, nil
}

func (j *Jenkins) DeleteProjectMember(membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error) {
	oldProjectRole, err := j.GetProjectRole(GetProjectRoleName(membership.ProjectId, membership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = oldProjectRole.UnAssignRole(membership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}

	oldPipelineRole, err := j.GetProjectRole(GetPipelineRoleName(membership.ProjectId, membership.Role))
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	err = oldPipelineRole.UnAssignRole(membership.Username)
	if err != nil {
		return nil, restful.NewError(GetJenkinsStatusCode(err), err.Error())
	}
	return membership, nil
}

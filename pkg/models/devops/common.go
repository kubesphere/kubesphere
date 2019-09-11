/*
Copyright 2019 The KubeSphere Authors.
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
	"fmt"
	"github.com/fatih/structs"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"kubesphere.io/kubesphere/pkg/utils/stringutils"
)

func GetColumnsFromStruct(s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = stringutils.CamelCaseToUnderscore(name)
	}
	return names
}

func GetColumnsFromStructWithPrefix(prefix string, s interface{}) []string {
	names := structs.Names(s)
	for i, name := range names {
		names[i] = WithPrefix(prefix, stringutils.CamelCaseToUnderscore(name))
	}
	return names
}

func WithPrefix(prefix, str string) string {
	return prefix + "." + str
}

const (
	StatusActive     = "active"
	StatusDeleted    = "deleted"
	StatusDeleting   = "deleting"
	StatusFailed     = "failed"
	StatusPending    = "pending"
	StatusWorking    = "working"
	StatusSuccessful = "successful"
)

const (
	StatusColumn     = "status"
	StatusTimeColumn = "status_time"
)

const (
	VisibilityPrivate = "private"
	VisibilityPublic  = "public"
)

const (
	KS_ADMIN = "admin"
)

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
)

type Role struct {
	Name        string `json:"name" description:"role's name e.g. owner'"`
	Description string `json:"description" description:"role 's description'"`
}

var DefaultRoles = []*Role{
	{
		Name:        ProjectOwner,
		Description: "Owner have access to do all the operations of a DevOps project and own the highest permissions as well.",
	},
	{
		Name:        ProjectMaintainer,
		Description: "Maintainer have access to manage pipeline and credential configuration in a DevOps project.",
	},
	{
		Name:        ProjectDeveloper,
		Description: "Developer is able to view and trigger the pipeline.",
	},
	{
		Name:        ProjectReporter,
		Description: "Reporter is only allowed to view the status of the pipeline.",
	},
}

var AllRoleSlice = []string{ProjectDeveloper, ProjectReporter, ProjectMaintainer, ProjectOwner}

var JenkinsOwnerProjectPermissionIds = &gojenkins.ProjectPermissionIds{
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

var JenkinsProjectPermissionMap = map[string]gojenkins.ProjectPermissionIds{
	ProjectOwner: {
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
	ProjectMaintainer: {
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
	ProjectDeveloper: {
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
	ProjectReporter: {
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

var JenkinsPipelinePermissionMap = map[string]gojenkins.ProjectPermissionIds{
	ProjectOwner: {
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
	ProjectMaintainer: {
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
	ProjectDeveloper: {
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
	ProjectReporter: {
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

func CheckProjectUserInRole(username, projectId string, roles []string) error {
	if username == KS_ADMIN {
		return nil
	}
	dbconn, err := client.ClientSets().MySQL()
	if err != nil {
		if _, ok := err.(client.ClientSetNotEnabledError); ok {
			klog.Error("mysql is not enabled")
		} else {
			klog.Error("error creating mysql client", err)
		}
		return nil
	}

	membership := &DevOpsProjectMembership{}
	err = dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId))).LoadOne(membership)
	if err != nil {
		return err
	}
	if !reflectutils.In(membership.Role, roles) {
		return fmt.Errorf("user [%s] in project [%s] role is not in %s", username, projectId, roles)
	}
	return nil
}

func GetProjectUserRole(username, projectId string) (string, error) {
	if username == KS_ADMIN {
		return ProjectOwner, nil
	}
	dbconn, err := client.ClientSets().MySQL()
	if err != nil {
		if _, ok := err.(client.ClientSetNotEnabledError); ok {
			klog.Error("mysql is not enabled")
		} else {
			klog.Error("error creating mysql client", err)
		}
		return "", err
	}
	membership := &DevOpsProjectMembership{}
	err = dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId))).LoadOne(membership)
	if err != nil {
		return "", err
	}

	return membership.Role, nil
}

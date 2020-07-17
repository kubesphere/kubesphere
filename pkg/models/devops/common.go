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
	"fmt"
	"github.com/fatih/structs"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
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

// define roles of DevOps
const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

const (
	JenkinsAllUserRoleName = "kubesphere-user"
	JenkinsAdminRoleName   = "admin"
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

// define the permission matrix of owner
var JenkinsOwnerProjectPermissionIds = &devops.ProjectPermissionIds{
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

// define the permission matrix of DevOps, including owner, maintainer, developer, reporter
var JenkinsProjectPermissionMap = map[string]devops.ProjectPermissionIds{
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

// define the permission matrix of pipeline, including owner, maintainer, developer, reporter
var JenkinsPipelinePermissionMap = map[string]devops.ProjectPermissionIds{
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

// get roleName of the project
func GetProjectRoleName(projectId, role string) string {
	return fmt.Sprintf("%s-%s-project", projectId, role)
}

// get roleName of the pipeline
func GetPipelineRoleName(projectId, role string) string {
	return fmt.Sprintf("%s-%s-pipeline", projectId, role)
}

// get pattern string of the project
func GetProjectRolePattern(projectId string) string {
	return fmt.Sprintf("^%s$", projectId)
}

// get pattern string of the project
func GetPipelineRolePattern(projectId string) string {
	return fmt.Sprintf("^%s/.*", projectId)
}

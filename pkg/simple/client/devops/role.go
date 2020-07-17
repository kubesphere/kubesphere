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

// define the id of global permission items
type GlobalPermissionIds struct {
	Administer              bool `json:"hudson.model.Hudson.Administer"`
	GlobalRead              bool `json:"hudson.model.Hudson.Read"`
	CredentialCreate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Create"`
	CredentialUpdate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Update"`
	CredentialView          bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.View"`
	CredentialDelete        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Delete"`
	CredentialManageDomains bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.ManageDomains"`
	SlaveCreate             bool `json:"hudson.model.Computer.Create"`
	SlaveConfigure          bool `json:"hudson.model.Computer.Configure"`
	SlaveDelete             bool `json:"hudson.model.Computer.Delete"`
	SlaveBuild              bool `json:"hudson.model.Computer.Build"`
	SlaveConnect            bool `json:"hudson.model.Computer.Connect"`
	SlaveDisconnect         bool `json:"hudson.model.Computer.Disconnect"`
	ItemBuild               bool `json:"hudson.model.Item.Build"`
	ItemCreate              bool `json:"hudson.model.Item.Create"`
	ItemRead                bool `json:"hudson.model.Item.Read"`
	ItemConfigure           bool `json:"hudson.model.Item.Configure"`
	ItemCancel              bool `json:"hudson.model.Item.Cancel"`
	ItemMove                bool `json:"hudson.model.Item.Move"`
	ItemDiscover            bool `json:"hudson.model.Item.Discover"`
	ItemWorkspace           bool `json:"hudson.model.Item.Workspace"`
	ItemDelete              bool `json:"hudson.model.Item.Delete"`
	RunUpdate               bool `json:"hudson.model.Run.Update"`
	RunDelete               bool `json:"hudson.model.Run.Delete"`
	ViewCreate              bool `json:"hudson.model.View.Create"`
	ViewConfigure           bool `json:"hudson.model.View.Configure"`
	ViewRead                bool `json:"hudson.model.View.Read"`
	ViewDelete              bool `json:"hudson.model.View.Delete"`
	SCMTag                  bool `json:"hudson.scm.SCM.Tag"`
}

// define the id of project permission items
type ProjectPermissionIds struct {
	CredentialCreate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Create"`
	CredentialUpdate        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Update"`
	CredentialView          bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.View"`
	CredentialDelete        bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.Delete"`
	CredentialManageDomains bool `json:"com.cloudbees.plugins.credentials.CredentialsProvider.ManageDomains"`
	ItemBuild               bool `json:"hudson.model.Item.Build"`
	ItemCreate              bool `json:"hudson.model.Item.Create"`
	ItemRead                bool `json:"hudson.model.Item.Read"`
	ItemConfigure           bool `json:"hudson.model.Item.Configure"`
	ItemCancel              bool `json:"hudson.model.Item.Cancel"`
	ItemMove                bool `json:"hudson.model.Item.Move"`
	ItemDiscover            bool `json:"hudson.model.Item.Discover"`
	ItemWorkspace           bool `json:"hudson.model.Item.Workspace"`
	ItemDelete              bool `json:"hudson.model.Item.Delete"`
	RunUpdate               bool `json:"hudson.model.Run.Update"`
	RunDelete               bool `json:"hudson.model.Run.Delete"`
	RunReplay               bool `json:"hudson.model.Run.Replay"`
	SCMTag                  bool `json:"hudson.scm.SCM.Tag"`
}

// describe the interface of DevOps to operator role
type RoleOperator interface {
	AddGlobalRole(roleName string, ids GlobalPermissionIds, overwrite bool) error
	GetGlobalRole(roleName string) (string, error)

	AddProjectRole(roleName string, pattern string, ids ProjectPermissionIds, overwrite bool) error
	DeleteProjectRoles(roleName ...string) error

	AssignProjectRole(roleName string, sid string) error
	UnAssignProjectRole(roleName string, sid string) error

	AssignGlobalRole(roleName string, sid string) error
	UnAssignGlobalRole(roleName string, sid string) error

	DeleteUserInProject(sid string) error
}

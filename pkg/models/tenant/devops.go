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
package tenant

import (
	"fmt"
	"github.com/gocraft/dbr"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/admin_jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/devops_mysql"
	"kubesphere.io/kubesphere/pkg/utils/reflectutils"
	"net/http"
	"sync"
)

const (
	ProjectOwner      = "owner"
	ProjectMaintainer = "maintainer"
	ProjectDeveloper  = "developer"
	ProjectReporter   = "reporter"
)

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
	ProjectOwner: gojenkins.ProjectPermissionIds{
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
	ProjectMaintainer: gojenkins.ProjectPermissionIds{
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
	ProjectDeveloper: gojenkins.ProjectPermissionIds{
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
	ProjectReporter: gojenkins.ProjectPermissionIds{
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
	ProjectOwner: gojenkins.ProjectPermissionIds{
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
	ProjectMaintainer: gojenkins.ProjectPermissionIds{
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
	ProjectDeveloper: gojenkins.ProjectPermissionIds{
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
	ProjectReporter: gojenkins.ProjectPermissionIds{
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

type DevOpsProjectRoleResponse struct {
	ProjectRole *gojenkins.ProjectRole
	Err         error
}

func CheckProjectUserInRole(username, projectId string, roles []string) error {
	if username == devops.KS_ADMIN {
		return nil
	}
	dbconn := devops_mysql.OpenDatabase()
	membership := &devops.DevOpsProjectMembership{}
	err := dbconn.Select(devops.DevOpsProjectMembershipColumns...).
		From(devops.DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(devops.DevOpsProjectMembershipUsernameColumn, username),
			db.Eq(devops.DevOpsProjectMembershipProjectIdColumn, projectId))).LoadOne(membership)
	if err != nil {
		return err
	}
	if !reflectutils.In(membership.Role, roles) {
		return fmt.Errorf("user [%s] in project [%s] role is not in %s", username, projectId, roles)
	}
	return nil
}

func ListDevopsProjects(workspace, username string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {

	dbconn := devops_mysql.OpenDatabase()

	query := dbconn.Select(devops.GetColumnsFromStructWithPrefix(devops.DevOpsProjectTableName, devops.DevOpsProject{})...).
		From(devops.DevOpsProjectTableName)
	var sqconditions []dbr.Builder

	sqconditions = append(sqconditions, db.Eq(devops.DevOpsProjectWorkSpaceColumn, workspace))

	switch username {
	case devops.KS_ADMIN:
	default:
		onCondition := fmt.Sprintf("%s = %s", devops.DevOpsProjectMembershipProjectIdColumn, devops.DevOpsProjectIdColumn)
		query.Join(devops.DevOpsProjectMembershipTableName, onCondition)
		sqconditions = append(sqconditions, db.Eq(devops.DevOpsProjectMembershipUsernameColumn, username))
		sqconditions = append(sqconditions, db.Eq(
			devops.DevOpsProjectMembershipTableName+"."+devops.StatusColumn, devops.StatusActive))
	}

	sqconditions = append(sqconditions, db.Eq(
		devops.DevOpsProjectTableName+"."+devops.StatusColumn, devops.StatusActive))
	if keyword := conditions.Match["keyword"]; keyword != "" {
		sqconditions = append(sqconditions, db.Like(devops.DevOpsProjectNameColumn, keyword))
	}
	projects := make([]*devops.DevOpsProject, 0)

	if len(sqconditions) > 0 {
		query.Where(db.And(sqconditions...))
	}
	switch orderBy {
	case "name":
		if reverse {
			query.OrderDesc(devops.DevOpsProjectNameColumn)
		} else {
			query.OrderAsc(devops.DevOpsProjectNameColumn)
		}
	default:
		if reverse {
			query.OrderAsc(devops.DevOpsProjectCreateTimeColumn)
		} else {
			query.OrderDesc(devops.DevOpsProjectCreateTimeColumn)
		}

	}
	query.Limit(uint64(limit))
	query.Offset(uint64(offset))
	_, err := query.Load(&projects)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err
	}
	count, err := query.Count()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err
	}

	result := make([]interface{}, 0)
	for _, v := range projects {
		result = append(result, v)
	}

	return &models.PageableResponse{Items: result, TotalCount: int(count)}, nil
}

func DeleteDevOpsProject(projectId, username string) (error, int) {
	err := CheckProjectUserInRole(username, projectId, []string{ProjectOwner})
	if err != nil {
		glog.Errorf("%+v", err)
		return err, http.StatusForbidden
	}
	gojenkins := admin_jenkins.Client()
	devopsdb := devops_mysql.OpenDatabase()
	_, err = gojenkins.DeleteJob(projectId)

	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		glog.Errorf("%+v", err)
		return err, utils.GetJenkinsStatusCode(err)
	}

	roleNames := make([]string, 0)
	for role := range JenkinsProjectPermissionMap {
		roleNames = append(roleNames, GetProjectRoleName(projectId, role))
		roleNames = append(roleNames, GetPipelineRoleName(projectId, role))
	}
	err = gojenkins.DeleteProjectRoles(roleNames...)
	if err != nil {
		glog.Errorf("%+v", err)
		return err, utils.GetJenkinsStatusCode(err)
	}
	_, err = devopsdb.DeleteFrom(devops.DevOpsProjectMembershipTableName).
		Where(db.Eq(devops.DevOpsProjectMembershipProjectIdColumn, projectId)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return err, http.StatusInternalServerError
	}
	_, err = devopsdb.Update(devops.DevOpsProjectTableName).
		Set(devops.StatusColumn, devops.StatusDeleted).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return err, http.StatusInternalServerError
	}
	project := &devops.DevOpsProject{}
	err = devopsdb.Select(devops.DevOpsProjectColumns...).
		From(devops.DevOpsProjectTableName).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).
		LoadOne(project)
	if err != nil {
		glog.Errorf("%+v", err)
		return err, http.StatusInternalServerError
	}
	return nil, http.StatusOK
}

func CreateDevopsProject(username string, workspace string, req *devops.DevOpsProject) (*devops.DevOpsProject, error, int) {

	jenkinsClient := admin_jenkins.Client()
	devopsdb := devops_mysql.OpenDatabase()
	project := devops.NewDevOpsProject(req.Name, req.Description, username, req.Extra, workspace)
	_, err := jenkinsClient.CreateFolder(project.ProjectId, project.Description)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}

	var addRoleCh = make(chan *DevOpsProjectRoleResponse, 8)
	var addRoleWg sync.WaitGroup
	for role, permission := range JenkinsProjectPermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission gojenkins.ProjectPermissionIds) {
			_, err := jenkinsClient.AddProjectRole(GetProjectRoleName(project.ProjectId, role),
				GetProjectRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	for role, permission := range JenkinsPipelinePermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission gojenkins.ProjectPermissionIds) {
			_, err := jenkinsClient.AddProjectRole(GetPipelineRoleName(project.ProjectId, role),
				GetPipelineRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	addRoleWg.Wait()
	close(addRoleCh)
	for addRoleResponse := range addRoleCh {
		if addRoleResponse.Err != nil {
			glog.Errorf("%+v", addRoleResponse.Err)
			return nil, addRoleResponse.Err, utils.GetJenkinsStatusCode(addRoleResponse.Err)
		}
	}

	globalRole, err := jenkinsClient.GetGlobalRole(devops.JenkinsAllUserRoleName)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}
	if globalRole == nil {
		_, err := jenkinsClient.AddGlobalRole(devops.JenkinsAllUserRoleName, gojenkins.GlobalPermissionIds{
			GlobalRead: true,
		}, true)
		if err != nil {
			glog.Error("failed to create jenkins global role")
			return nil, err, utils.GetJenkinsStatusCode(err)
		}
	}
	err = globalRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}

	projectRole, err := jenkinsClient.GetProjectRole(GetProjectRoleName(project.ProjectId, ProjectOwner))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}
	err = projectRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}

	pipelineRole, err := jenkinsClient.GetProjectRole(GetPipelineRoleName(project.ProjectId, ProjectOwner))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}
	err = pipelineRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, utils.GetJenkinsStatusCode(err)
	}
	_, err = devopsdb.InsertInto(devops.DevOpsProjectTableName).
		Columns(devops.DevOpsProjectColumns...).Record(project).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, http.StatusInternalServerError
	}

	projectMembership := devops.NewDevOpsProjectMemberShip(username, project.ProjectId, ProjectOwner, username)
	_, err = devopsdb.InsertInto(devops.DevOpsProjectMembershipTableName).
		Columns(devops.DevOpsProjectMembershipColumns...).Record(projectMembership).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, err, http.StatusInternalServerError
	}
	return project, nil, http.StatusOK
}

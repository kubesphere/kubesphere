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
	"github.com/emicklei/go-restful"
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
	"net/http"
	"sync"
)

type DevOpsProjectRoleResponse struct {
	ProjectRole *gojenkins.ProjectRole
	Err         error
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
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	count, err := query.Count()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	result := make([]interface{}, 0)
	for _, v := range projects {
		result = append(result, v)
	}

	return &models.PageableResponse{Items: result, TotalCount: int(count)}, nil
}

func DeleteDevOpsProject(projectId, username string) error {
	err := devops.CheckProjectUserInRole(username, projectId, []string{devops.ProjectOwner})
	if err != nil {
		glog.Errorf("%+v", err)
		return restful.NewError(http.StatusForbidden, err.Error())
	}
	gojenkins := admin_jenkins.Client()
	if gojenkins == nil {
		err := fmt.Errorf("could not connect to jenkins")
		glog.Error(err)
		return restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	devopsdb := devops_mysql.OpenDatabase()
	_, err = gojenkins.DeleteJob(projectId)

	if err != nil && utils.GetJenkinsStatusCode(err) != http.StatusNotFound {
		glog.Errorf("%+v", err)
		return restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	roleNames := make([]string, 0)
	for role := range devops.JenkinsProjectPermissionMap {
		roleNames = append(roleNames, devops.GetProjectRoleName(projectId, role))
		roleNames = append(roleNames, devops.GetPipelineRoleName(projectId, role))
	}
	err = gojenkins.DeleteProjectRoles(roleNames...)
	if err != nil {
		glog.Errorf("%+v", err)
		return restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	_, err = devopsdb.DeleteFrom(devops.DevOpsProjectMembershipTableName).
		Where(db.Eq(devops.DevOpsProjectMembershipProjectIdColumn, projectId)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	_, err = devopsdb.Update(devops.DevOpsProjectTableName).
		Set(devops.StatusColumn, devops.StatusDeleted).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	project := &devops.DevOpsProject{}
	err = devopsdb.Select(devops.DevOpsProjectColumns...).
		From(devops.DevOpsProjectTableName).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).
		LoadOne(project)
	if err != nil {
		glog.Errorf("%+v", err)
		return restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	return nil
}

func CreateDevopsProject(username string, workspace string, req *devops.DevOpsProject) (*devops.DevOpsProject, error) {

	jenkinsClient := admin_jenkins.Client()
	if jenkinsClient == nil {
		err := fmt.Errorf("could not connect to jenkins")
		glog.Error(err)
		return nil, restful.NewError(http.StatusServiceUnavailable, err.Error())
	}
	devopsdb := devops_mysql.OpenDatabase()
	project := devops.NewDevOpsProject(req.Name, req.Description, username, req.Extra, workspace)
	_, err := jenkinsClient.CreateFolder(project.ProjectId, project.Description)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	var addRoleCh = make(chan *DevOpsProjectRoleResponse, 8)
	var addRoleWg sync.WaitGroup
	for role, permission := range devops.JenkinsProjectPermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission gojenkins.ProjectPermissionIds) {
			_, err := jenkinsClient.AddProjectRole(devops.GetProjectRoleName(project.ProjectId, role),
				devops.GetProjectRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	for role, permission := range devops.JenkinsPipelinePermissionMap {
		addRoleWg.Add(1)
		go func(role string, permission gojenkins.ProjectPermissionIds) {
			_, err := jenkinsClient.AddProjectRole(devops.GetPipelineRoleName(project.ProjectId, role),
				devops.GetPipelineRolePattern(project.ProjectId), permission, true)
			addRoleCh <- &DevOpsProjectRoleResponse{nil, err}
			addRoleWg.Done()
		}(role, permission)
	}
	addRoleWg.Wait()
	close(addRoleCh)
	for addRoleResponse := range addRoleCh {
		if addRoleResponse.Err != nil {
			glog.Errorf("%+v", addRoleResponse.Err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(addRoleResponse.Err), addRoleResponse.Err.Error())
		}
	}

	globalRole, err := jenkinsClient.GetGlobalRole(devops.JenkinsAllUserRoleName)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	if globalRole == nil {
		_, err := jenkinsClient.AddGlobalRole(devops.JenkinsAllUserRoleName, gojenkins.GlobalPermissionIds{
			GlobalRead: true,
		}, true)
		if err != nil {
			glog.Error("failed to create jenkins global role")
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
	}
	err = globalRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	projectRole, err := jenkinsClient.GetProjectRole(devops.GetProjectRoleName(project.ProjectId, devops.ProjectOwner))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = projectRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	pipelineRole, err := jenkinsClient.GetProjectRole(devops.GetPipelineRoleName(project.ProjectId, devops.ProjectOwner))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = pipelineRole.AssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	_, err = devopsdb.InsertInto(devops.DevOpsProjectTableName).
		Columns(devops.DevOpsProjectColumns...).Record(project).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	projectMembership := devops.NewDevOpsProjectMemberShip(username, project.ProjectId, devops.ProjectOwner, username)
	_, err = devopsdb.InsertInto(devops.DevOpsProjectMembershipTableName).
		Columns(devops.DevOpsProjectMembershipColumns...).Record(projectMembership).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return project, nil
}

func GetUserDevopsSimpleRules(username, projectId string) ([]models.SimpleRule, error) {
	role, err := devops.GetProjectUserRole(username, projectId)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusForbidden, err.Error())
	}
	return GetDevopsRoleSimpleRules(role), nil
}

func GetDevopsRoleSimpleRules(role string) []models.SimpleRule {
	var rules []models.SimpleRule

	switch role {
	case "developer":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"view", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "owner":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"edit", "view", "delete"}},
		}
		break
	case "maintainer":
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "reporter":
		fallthrough
	default:
		rules = []models.SimpleRule{
			{Name: "pipelines", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	}
	return rules
}

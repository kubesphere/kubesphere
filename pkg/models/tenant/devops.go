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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api/devops/v1alpha2"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/models/devops"
	"kubesphere.io/kubesphere/pkg/models/iam"
	"kubesphere.io/kubesphere/pkg/server/params"
	dsClient "kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"net/http"
)

type DevOpsProjectOperator interface {
	ListDevOpsProjects(workspace, username string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error)
	CreateDevOpsProject(username string, workspace string, req *v1alpha2.DevOpsProject) (*v1alpha2.DevOpsProject, error)
	GetDevOpsProjectsCount(username string) (uint32, error)
	DeleteDevOpsProject(projectId, username string) error
	GetUserDevOpsSimpleRules(username, projectId string) ([]models.SimpleRule, error)
}

type devopsProjectOperator struct {
	ksProjectOperator devops.ProjectOperator
	db                *mysql.Database
	dsProject         dsClient.ProjectOperator
}

func (o *devopsProjectOperator) ListDevOpsProjects(workspace, username string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {

	query := o.db.Select(devops.GetColumnsFromStructWithPrefix(devops.DevOpsProjectTableName, v1alpha2.DevOpsProject{})...).
		From(devops.DevOpsProjectTableName)
	var sqconditions []dbr.Builder

	sqconditions = append(sqconditions, db.Eq(devops.DevOpsProjectWorkSpaceColumn, workspace))

	switch username {
	case devops.KS_ADMIN:
	default:
		onCondition := fmt.Sprintf("%s = %s", devops.ProjectMembershipProjectIdColumn, devops.DevOpsProjectIdColumn)
		query.Join(devops.ProjectMembershipTableName, onCondition)
		sqconditions = append(sqconditions, db.Eq(devops.ProjectMembershipUsernameColumn, username))
		sqconditions = append(sqconditions, db.Eq(
			devops.ProjectMembershipTableName+"."+devops.StatusColumn, devops.StatusActive))
	}

	sqconditions = append(sqconditions, db.Eq(
		devops.DevOpsProjectTableName+"."+devops.StatusColumn, devops.StatusActive))
	if keyword := conditions.Match["keyword"]; keyword != "" {
		sqconditions = append(sqconditions, db.Like(devops.DevOpsProjectNameColumn, keyword))
	}
	projects := make([]*v1alpha2.DevOpsProject, 0)

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
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	count, err := query.Count()
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	result := make([]interface{}, 0)
	for _, v := range projects {
		result = append(result, v)
	}

	return &models.PageableResponse{Items: result, TotalCount: int(count)}, nil
}

func (o *devopsProjectOperator) GetDevOpsProjectsCount(username string) (uint32, error) {

	query := o.db.Select(devops.GetColumnsFromStructWithPrefix(devops.DevOpsProjectTableName, v1alpha2.DevOpsProject{})...).
		From(devops.DevOpsProjectTableName)
	var sqconditions []dbr.Builder

	if username != devops.KS_ADMIN {
		onCondition := fmt.Sprintf("%s = %s", devops.ProjectMembershipProjectIdColumn, devops.DevOpsProjectIdColumn)
		query.Join(devops.ProjectMembershipTableName, onCondition)
		sqconditions = append(sqconditions, db.Eq(devops.ProjectMembershipUsernameColumn, username))
		sqconditions = append(sqconditions, db.Eq(
			devops.ProjectMembershipTableName+"."+devops.StatusColumn, devops.StatusActive))
	}

	sqconditions = append(sqconditions, db.Eq(
		devops.DevOpsProjectTableName+"."+devops.StatusColumn, devops.StatusActive))
	if len(sqconditions) > 0 {
		query.Where(db.And(sqconditions...))
	}
	count, err := query.Count()
	if err != nil {
		klog.Errorf("%+v", err)
		return 0, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return count, nil
}

func (o *devopsProjectOperator) DeleteDevOpsProject(projectId, username string) error {
	err := o.ksProjectOperator.CheckProjectUserInRole(username, projectId, []string{dsClient.ProjectOwner})
	if err != nil {
		klog.Errorf("%+v", err)
		return restful.NewError(http.StatusForbidden, err.Error())
	}

	err = o.dsProject.DeleteDevOpsProject(projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	_, err = o.db.DeleteFrom(devops.ProjectMembershipTableName).
		Where(db.Eq(devops.ProjectMembershipProjectIdColumn, projectId)).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	_, err = o.db.Update(devops.DevOpsProjectTableName).
		Set(devops.StatusColumn, devops.StatusDeleted).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	project := &v1alpha2.DevOpsProject{}
	err = o.db.Select(devops.DevOpsProjectColumns...).
		From(devops.DevOpsProjectTableName).
		Where(db.Eq(devops.DevOpsProjectIdColumn, projectId)).
		LoadOne(project)
	if err != nil {
		klog.Errorf("%+v", err)
		return err
	}
	return nil
}

func (o *devopsProjectOperator) CreateDevOpsProject(username string, workspace string, req *v1alpha2.DevOpsProject) (*v1alpha2.DevOpsProject, error) {

	project := devops.NewDevOpsProject(req.Name, req.Description, username, req.Extra, workspace)
	_, err := o.dsProject.CreateDevOpsProject(username, project)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	_, err = o.db.InsertInto(devops.DevOpsProjectTableName).
		Columns(devops.DevOpsProjectColumns...).Record(project).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	projectMembership := devops.NewDevOpsProjectMemberShip(username, project.ProjectId, dsClient.ProjectOwner, username)
	_, err = o.db.InsertInto(devops.ProjectMembershipTableName).
		Columns(devops.ProjectMembershipColumns...).Record(projectMembership).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return project, nil
}

<<<<<<< HEAD
func (o *devopsProjectOperator) GetUserDevOpsSimpleRules(username, projectId string) ([]models.SimpleRule, error) {
=======
func GetUserDevopsSimpleRules(username, projectId string) ([]iam.SimpleRule, error) {
>>>>>>> 71849f028f8afb6f270f2e8ec07128e2e2e1cfa2
	role, err := devops.GetProjectUserRole(username, projectId)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusForbidden, err.Error())
	}
	return GetDevopsRoleSimpleRules(role), nil
}

func GetDevopsRoleSimpleRules(role string) []iam.SimpleRule {
	var rules []iam.SimpleRule

	switch role {
	case "developer":
		rules = []iam.SimpleRule{
			{Name: "pipelines", Actions: []string{"view", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	case "owner":
		rules = []iam.SimpleRule{
			{Name: "pipelines", Actions: []string{"create", "edit", "view", "delete", "trigger"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "credentials", Actions: []string{"create", "edit", "view", "delete"}},
			{Name: "devops", Actions: []string{"edit", "view", "delete"}},
		}
		break
	case "maintainer":
		rules = []iam.SimpleRule{
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
		rules = []iam.SimpleRule{
			{Name: "pipelines", Actions: []string{"view"}},
			{Name: "roles", Actions: []string{"view"}},
			{Name: "members", Actions: []string{"view"}},
			{Name: "devops", Actions: []string{"view"}},
		}
		break
	}
	return rules
}

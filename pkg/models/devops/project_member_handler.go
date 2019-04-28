/*
Copyright 2018 The KubeSphere Authors.
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
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/gojenkins"
	"kubesphere.io/kubesphere/pkg/gojenkins/utils"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/params"
	"kubesphere.io/kubesphere/pkg/simple/client/admin_jenkins"
	"kubesphere.io/kubesphere/pkg/simple/client/devops_mysql"
)

func GetProjectMembers(projectId string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {
	dbconn := devops_mysql.OpenDatabase()
	memberships := make([]*DevOpsProjectMembership, 0)
	var sqconditions []dbr.Builder
	sqconditions = append(sqconditions, db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId))
	if keyword := conditions.Match["keyword"]; keyword != "" {
		sqconditions = append(sqconditions, db.Like(DevOpsProjectMembershipUsernameColumn, keyword))
	}
	query := dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName)
	switch orderBy {
	case "name":
		if reverse {
			query.OrderDesc(DevOpsProjectMembershipUsernameColumn)
		} else {
			query.OrderAsc(DevOpsProjectMembershipUsernameColumn)
		}
	default:
		if reverse {
			query.OrderDesc(DevOpsProjectMembershipRoleColumn)
		} else {
			query.OrderAsc(DevOpsProjectMembershipRoleColumn)
		}
	}
	query.Limit(uint64(limit))
	query.Offset(uint64(offset))
	if len(sqconditions) > 1 {
		query.Where(db.And(sqconditions...))
	} else {
		query.Where(sqconditions[0])
	}
	_, err := query.Load(&memberships)
	if err != nil && err != dbr.ErrNotFound {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	count, err := query.Count()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	result := make([]interface{}, 0)
	for _, v := range memberships {
		result = append(result, v)
	}

	return &models.PageableResponse{Items: result, TotalCount: int(count)}, nil
}

func GetProjectMember(projectId, username string) (*DevOpsProjectMembership, error) {
	dbconn := devops_mysql.OpenDatabase()
	member := &DevOpsProjectMembership{}
	err := dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
			db.Eq(DevOpsProjectMembershipUsernameColumn, username))).
		LoadOne(&member)
	if err != nil && err != dbr.ErrNotFound {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	if err == dbr.ErrNotFound {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusNotFound, err.Error())
	}
	return member, nil
}

func AddProjectMember(projectId, operator string, member *DevOpsProjectMembership) (*DevOpsProjectMembership, error) {
	dbconn := devops_mysql.OpenDatabase()
	jenkinsClinet := admin_jenkins.Client()

	membership := &DevOpsProjectMembership{}
	err := dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, member.Username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId))).LoadOne(membership)
	// if user could be founded in db, user have been added to project
	if err == nil {
		err = fmt.Errorf("user [%s] have been added to project", member.Username)
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}

	if err != nil && err != db.ErrNotFound {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	globalRole, err := jenkinsClinet.GetGlobalRole(JenkinsAllUserRoleName)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	if globalRole == nil {
		_, err := jenkinsClinet.AddGlobalRole(JenkinsAllUserRoleName, gojenkins.GlobalPermissionIds{
			GlobalRead: true,
		}, true)
		if err != nil {
			glog.Errorf("failed to create jenkins global role %+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
	}
	err = globalRole.AssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	projectRole, err := jenkinsClinet.GetProjectRole(GetProjectRoleName(projectId, member.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = projectRole.AssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	pipelineRole, err := jenkinsClinet.GetProjectRole(GetPipelineRoleName(projectId, member.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = pipelineRole.AssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	projectMembership := NewDevOpsProjectMemberShip(member.Username, projectId, member.Role, operator)
	_, err = dbconn.
		InsertInto(DevOpsProjectMembershipTableName).
		Columns(DevOpsProjectMembershipColumns...).
		Record(projectMembership).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		err = projectRole.UnAssignRole(member.Username)
		if err != nil {
			glog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		err = pipelineRole.UnAssignRole(member.Username)
		if err != nil {
			glog.Errorf("%+v", err)
			return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
		}
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return projectMembership, nil
}

func UpdateProjectMember(projectId, operator string, member *DevOpsProjectMembership) (*DevOpsProjectMembership, error) {
	dbconn := devops_mysql.OpenDatabase()
	jenkinsClinet := admin_jenkins.Client()
	oldMembership := &DevOpsProjectMembership{}
	err := dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, member.Username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(oldMembership)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}

	oldProjectRole, err := jenkinsClinet.GetProjectRole(GetProjectRoleName(projectId, oldMembership.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = oldProjectRole.UnAssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	oldPipelineRole, err := jenkinsClinet.GetProjectRole(GetPipelineRoleName(projectId, oldMembership.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = oldPipelineRole.UnAssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	projectRole, err := jenkinsClinet.GetProjectRole(GetProjectRoleName(projectId, member.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = projectRole.AssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	pipelineRole, err := jenkinsClinet.GetProjectRole(GetPipelineRoleName(projectId, member.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = pipelineRole.AssignRole(member.Username)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	_, err = dbconn.Update(DevOpsProjectMembershipTableName).
		Set(DevOpsProjectMembershipRoleColumn, member.Role).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
			db.Eq(DevOpsProjectMembershipUsernameColumn, member.Username),
		)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	responseMembership := &DevOpsProjectMembership{}
	err = dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, member.Username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(responseMembership)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return responseMembership, nil
}

func DeleteProjectMember(projectId, username string) (string, error) {
	dbconn := devops_mysql.OpenDatabase()
	jenkinsClient := admin_jenkins.Client()
	oldMembership := &DevOpsProjectMembership{}
	err := dbconn.Select(DevOpsProjectMembershipColumns...).
		From(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipUsernameColumn, username),
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(oldMembership)
	if err != nil && err != db.ErrNotFound {
		glog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusInternalServerError, err.Error())
	}
	if err == db.ErrNotFound {
		glog.Warningf("user [%s] not found in project", username)
		return username, nil
	}
	if oldMembership.Role == ProjectOwner {
		count, err := dbconn.Select(DevOpsProjectMembershipProjectIdColumn).
			From(DevOpsProjectMembershipTableName).
			Where(db.And(
				db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
				db.Eq(DevOpsProjectMembershipRoleColumn, ProjectOwner))).Count()
		if err != nil {
			glog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		if count == 1 {
			err = fmt.Errorf("project must has at least one admin")
			glog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	}
	oldProjectRole, err := jenkinsClient.GetProjectRole(GetProjectRoleName(projectId, oldMembership.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = oldProjectRole.UnAssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	oldPipelineRole, err := jenkinsClient.GetProjectRole(GetPipelineRoleName(projectId, oldMembership.Role))
	if err != nil {
		glog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}
	err = oldPipelineRole.UnAssignRole(username)
	if err != nil {
		glog.Errorf("%+v", err)
		return "", restful.NewError(utils.GetJenkinsStatusCode(err), err.Error())
	}

	_, err = dbconn.DeleteFrom(DevOpsProjectMembershipTableName).
		Where(db.And(
			db.Eq(DevOpsProjectMembershipProjectIdColumn, projectId),
			db.Eq(DevOpsProjectMembershipUsernameColumn, username),
		)).Exec()
	if err != nil {
		glog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return username, nil
}

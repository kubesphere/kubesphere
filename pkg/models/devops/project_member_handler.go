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
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client/devops"
	"kubesphere.io/kubesphere/pkg/simple/client/mysql"
	"net/http"

	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/models"
	"kubesphere.io/kubesphere/pkg/server/params"
)

type ProjectMemberOperator interface {
	GetProjectMembers(projectId string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error)
	GetProjectMember(projectId, username string) (*devops.DevOpsProjectMembership, error)
	AddProjectMember(projectId string, membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error)
	UpdateProjectMember(projectId string, membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error)
	DeleteProjectMember(projectId, username string) (string, error)
}
type projectMemberOperator struct {
	db                    *mysql.Database
	projectMemberOperator devops.ProjectMemberOperator
}

func NewProjectMemberOperator() ProjectMemberOperator {
	return &projectMemberOperator{}
}

func (o *projectMemberOperator) GetProjectMembers(projectId string, conditions *params.Conditions, orderBy string, reverse bool, limit int, offset int) (*models.PageableResponse, error) {

	memberships := make([]*devops.DevOpsProjectMembership, 0)
	var sqconditions []dbr.Builder
	sqconditions = append(sqconditions, db.Eq(ProjectMembershipProjectIdColumn, projectId))
	if keyword := conditions.Match["keyword"]; keyword != "" {
		sqconditions = append(sqconditions, db.Like(ProjectMembershipUsernameColumn, keyword))
	}
	query := *o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName)
	switch orderBy {
	case "name":
		if reverse {
			query.OrderDesc(ProjectMembershipUsernameColumn)
		} else {
			query.OrderAsc(ProjectMembershipUsernameColumn)
		}
	default:
		if reverse {
			query.OrderDesc(ProjectMembershipRoleColumn)
		} else {
			query.OrderAsc(ProjectMembershipRoleColumn)
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
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	count, err := query.Count()
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	result := make([]interface{}, 0)
	for _, v := range memberships {
		result = append(result, v)
	}

	return &models.PageableResponse{Items: result, TotalCount: int(count)}, nil
}

func (o *projectMemberOperator) GetProjectMember(projectId, username string) (*devops.DevOpsProjectMembership, error) {

	member := &devops.DevOpsProjectMembership{}
	err := o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName).
		Where(db.And(db.Eq(ProjectMembershipProjectIdColumn, projectId),
			db.Eq(ProjectMembershipUsernameColumn, username))).
		LoadOne(&member)
	if err != nil && err != dbr.ErrNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	if err == dbr.ErrNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusNotFound, err.Error())
	}
	return member, nil
}

func (o *projectMemberOperator) AddProjectMember(projectId string, membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error) {

	dbmembership := &devops.DevOpsProjectMembership{}
	err := o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName).
		Where(db.And(
			db.Eq(ProjectMembershipUsernameColumn, membership.Username),
			db.Eq(ProjectMembershipProjectIdColumn, projectId))).LoadOne(dbmembership)
	// if user could be founded in db, user have been added to project
	if err == nil {
		err = fmt.Errorf("user [%s] have been added to project", membership.Username)
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}

	if err != db.ErrNotFound {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	_, err = o.projectMemberOperator.AddProjectMember(membership)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, err
	}
	projectMembership := NewDevOpsProjectMemberShip(membership.Username, projectId, membership.Role, membership.GrantBy)
	_, err = o.db.
		InsertInto(ProjectMembershipTableName).
		Columns(ProjectMembershipColumns...).
		Record(projectMembership).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		_, err = o.projectMemberOperator.DeleteProjectMember(membership)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil, err
		}
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return projectMembership, nil
}

func (o *projectMemberOperator) UpdateProjectMember(projectId string, membership *devops.DevOpsProjectMembership) (*devops.DevOpsProjectMembership, error) {

	oldMembership := &devops.DevOpsProjectMembership{}
	err := o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName).
		Where(db.And(
			db.Eq(ProjectMembershipUsernameColumn, membership.Username),
			db.Eq(ProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(oldMembership)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusBadRequest, err.Error())
	}

	_, err = o.projectMemberOperator.UpdateProjectMember(oldMembership, membership)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, err
	}
	_, err = o.db.Update(ProjectMembershipTableName).
		Set(ProjectMembershipRoleColumn, membership.Role).
		Where(db.And(
			db.Eq(ProjectMembershipProjectIdColumn, projectId),
			db.Eq(ProjectMembershipUsernameColumn, membership.Username),
		)).Exec()

	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}

	responseMembership := &devops.DevOpsProjectMembership{}
	err = o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName).
		Where(db.And(
			db.Eq(ProjectMembershipUsernameColumn, membership.Username),
			db.Eq(ProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(responseMembership)
	if err != nil {
		klog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return responseMembership, nil
}

func (o *projectMemberOperator) DeleteProjectMember(projectId, username string) (string, error) {

	oldMembership := &devops.DevOpsProjectMembership{}
	err := o.db.Select(ProjectMembershipColumns...).
		From(ProjectMembershipTableName).
		Where(db.And(
			db.Eq(ProjectMembershipUsernameColumn, username),
			db.Eq(ProjectMembershipProjectIdColumn, projectId),
		)).LoadOne(oldMembership)
	if err != nil {
		if err != db.ErrNotFound {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		} else if err == db.ErrNotFound {
			klog.Warningf("user [%s] not found in project", username)
			return username, nil
		}
	}

	if oldMembership.Role == devops.ProjectOwner {
		count, err := o.db.Select(ProjectMembershipProjectIdColumn).
			From(ProjectMembershipTableName).
			Where(db.And(
				db.Eq(ProjectMembershipProjectIdColumn, projectId),
				db.Eq(ProjectMembershipRoleColumn, devops.ProjectOwner))).Count()
		if err != nil {
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusInternalServerError, err.Error())
		}
		if count == 1 {
			err = fmt.Errorf("project must has at least one admin")
			klog.Errorf("%+v", err)
			return "", restful.NewError(http.StatusBadRequest, err.Error())
		}
	}

	_, err = o.projectMemberOperator.DeleteProjectMember(oldMembership)
	if err != nil {
		klog.Error(err)
		return "", err
	}

	_, err = o.db.DeleteFrom(ProjectMembershipTableName).
		Where(db.And(
			db.Eq(ProjectMembershipProjectIdColumn, projectId),
			db.Eq(ProjectMembershipUsernameColumn, username),
		)).Exec()
	if err != nil {
		klog.Errorf("%+v", err)
		return "", restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return username, nil
}

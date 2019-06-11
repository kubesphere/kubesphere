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
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"github.com/golang/glog"
	"kubesphere.io/kubesphere/pkg/db"
	"kubesphere.io/kubesphere/pkg/simple/client/devops_mysql"
	"net/http"
)

func GetProject(projectId string) (*DevOpsProject, error) {
	dbconn := devops_mysql.OpenDatabase()
	project := &DevOpsProject{}
	err := dbconn.Select(DevOpsProjectColumns...).
		From(DevOpsProjectTableName).
		Where(db.Eq(DevOpsProjectIdColumn, projectId)).
		LoadOne(project)
	if err != nil && err != dbr.ErrNotFound {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	if err == dbr.ErrNotFound {
		glog.Errorf("%+v", err)

		return nil, restful.NewError(http.StatusNotFound, err.Error())
	}
	return project, nil
}

func UpdateProject(project *DevOpsProject) (*DevOpsProject, error) {
	dbconn := devops_mysql.OpenDatabase()
	query := dbconn.Update(DevOpsProjectTableName)
	if !govalidator.IsNull(project.Description) {
		query.Set(DevOpsProjectDescriptionColumn, project.Description)
	}
	if !govalidator.IsNull(project.Extra) {
		query.Set(DevOpsProjectExtraColumn, project.Extra)
	}
	if !govalidator.IsNull(project.Name) {
		query.Set(DevOpsProjectNameColumn, project.Extra)
	}
	if len(query.UpdateStmt.Value) > 0 {
		_, err := query.
			Where(db.Eq(DevOpsProjectIdColumn, project.ProjectId)).Exec()

		if err != nil {
			glog.Errorf("%+v", err)

			return nil, restful.NewError(http.StatusInternalServerError, err.Error())
		}
	}
	newProject := &DevOpsProject{}
	err := dbconn.Select(DevOpsProjectColumns...).
		From(DevOpsProjectTableName).
		Where(db.Eq(DevOpsProjectIdColumn, project.ProjectId)).
		LoadOne(newProject)
	if err != nil {
		glog.Errorf("%+v", err)
		return nil, restful.NewError(http.StatusInternalServerError, err.Error())
	}
	return newProject, nil
}

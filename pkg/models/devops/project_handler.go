package devops

import (
	"github.com/asaskevich/govalidator"
	"github.com/emicklei/go-restful"
	"github.com/gocraft/dbr"
	"k8s.io/klog/glog"
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

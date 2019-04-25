package devops

import (
	"kubesphere.io/kubesphere/pkg/utils/idutils"
	"time"
)

var DevOpsProjectColumns = GetColumnsFromStruct(&DevOpsProject{})

const (
	DevOpsProjectTableName         = "project"
	DevOpsProjectPrefix            = "project-"
	DevOpsProjectDescriptionColumn = "description"
	DevOpsProjectIdColumn          = "project.project_id"
	DevOpsProjectNameColumn        = "project.name"
	DevOpsProjectExtraColumn       = "project.extra"
	DevOpsProjectWorkSpaceColumn   = "project.workspace"
	DevOpsProjectCreateTimeColumn  = "project.create_time"
)

type DevOpsProject struct {
	ProjectId   string    `json:"project_id" db:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Creator     string    `json:"creator"`
	CreateTime  time.Time `json:"create_time"`
	Status      string    `json:"status"`
	Visibility  string    `json:"visibility"`
	Extra       string    `json:"extra"`
	Workspace   string    `json:"workspace"`
}

func NewDevOpsProject(name, description, creator, extra, workspace string) *DevOpsProject {
	return &DevOpsProject{
		ProjectId:   idutils.GetUuid(DevOpsProjectPrefix),
		Name:        name,
		Description: description,
		Creator:     creator,
		CreateTime:  time.Now(),
		Status:      StatusActive,
		Visibility:  VisibilityPrivate,
		Extra:       extra,
		Workspace:   workspace,
	}
}

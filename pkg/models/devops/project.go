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

type PageableDevOpsProject struct {
	Items      []*DevOpsProject `json:"items"`
	TotalCount int              `json:"total_count"`
}

type DevOpsProject struct {
	ProjectId   string    `json:"project_id" db:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Creator     string    `json:"creator"`
	CreateTime  time.Time `json:"create_time"`
	Status      string    `json:"status"`
	Visibility  string    `json:"visibility,omitempty"`
	Extra       string    `json:"extra,omitempty"`
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

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

package v1alpha2

import "time"

type PageableDevOpsProject struct {
	Items      []*DevOpsProject `json:"items"`
	TotalCount int              `json:"total_count"`
}

type DevOpsProject struct {
	ProjectId   string    `json:"project_id" db:"project_id" description:"ProjectId must be unique within a workspace, it is generated by kubesphere."`
	Name        string    `json:"name" description:"DevOps Projects's Name"`
	Description string    `json:"description,omitempty" description:"DevOps Projects's Description, used to describe the DevOps Project"`
	Creator     string    `json:"creator" description:"Creator's username"`
	CreateTime  time.Time `json:"create_time" description:"DevOps Project's Creation time"`
	Status      string    `json:"status" description:"DevOps project's status. e.g. active"`
	Visibility  string    `json:"visibility,omitempty" description:"Deprecated Field"`
	Extra       string    `json:"extra,omitempty" description:"Internal Use"`
	Workspace   string    `json:"workspace" description:"The workspace to which the devops project belongs"`
}

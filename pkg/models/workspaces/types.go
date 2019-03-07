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
package workspaces

import "time"

type Workspace struct {
	Group          `json:",inline"`
	Admin          string   `json:"admin,omitempty"`
	Namespaces     []string `json:"namespaces"`
	DevopsProjects []string `json:"devops_projects"`
}

type UserInvite struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

type Group struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Gid         string   `json:"gid"`
	Members     []string `json:"members"`
	Logo        string   `json:"logo"`
	Creator     string   `json:"creator"`
	CreateTime  string   `json:"create_time"`
	ChildGroups []string `json:"child_groups,omitempty"`
	Description string   `json:"description"`
}

func (g Group) GetCreateTime() (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", g.CreateTime)
}

type WorkspaceDPBinding struct {
	Workspace     string `gorm:"primary_key"`
	DevOpsProject string `gorm:"primary_key"`
}

type DevopsProject struct {
	ProjectId   *string    `json:"project_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Creator     string     `json:"creator"`
	CreateTime  *time.Time `json:"create_time,omitempty"`
	Status      *string    `json:"status"`
	Visibility  *string    `json:"visibility,omitempty"`
}

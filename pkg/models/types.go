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
package models

import (
	v12 "k8s.io/api/core/v1"
	"time"

	"k8s.io/api/rbac/v1"
)

type PageableResponse struct {
	Items      []interface{} `json:"items"`
	TotalCount int           `json:"total_count"`
}

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

type Action struct {
	Name  string          `json:"name"`
	Rules []v1.PolicyRule `json:"rules"`
}

type Rule struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type SimpleRule struct {
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type User struct {
	Username string `json:"username"`
	//UID      string   `json:"uid"`
	Groups          []string `json:"groups,omitempty"`
	Password        string   `json:"password,omitempty"`
	CurrentPassword string   `json:"current_password,omitempty"`
	//Extra    map[string]interface{} `json:"extra"`
	AvatarUrl      string                  `json:"avatar_url"`
	Description    string                  `json:"description"`
	Email          string                  `json:"email"`
	LastLoginTime  string                  `json:"last_login_time"`
	Status         int                     `json:"status"`
	ClusterRole    string                  `json:"cluster_role"`
	ClusterRules   []SimpleRule            `json:"cluster_rules"`
	Roles          map[string]string       `json:"roles,omitempty"`
	Rules          map[string][]SimpleRule `json:"rules,omitempty"`
	Role           string                  `json:"role,omitempty"`
	RoleBinding    string                  `json:"role_binding,omitempty"`
	Lang           string                  `json:"lang,omitempty"`
	WorkspaceRoles map[string]string       `json:"workspace_roles,omitempty"`
	WorkspaceRole  string                  `json:"workspace_role,omitempty"`
	WorkspaceRules map[string][]SimpleRule `json:"workspace_rules,omitempty"`
}

type Group struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Gid         string   `json:"gid"`
	Members     []string `json:"members"`
	Logo        string   `json:"logo"`
	Creator     string   `json:"creator"`
	CreateTime  string   `json:"create_time"`
	ChildGroups []string `json:"child_groups"`
	Description string   `json:"description"`
}

type Component struct {
	Name            string      `json:"name"`
	Namespace       string      `json:"namespace"`
	SelfLink        string      `json:"selfLink"`
	Label           interface{} `json:"label"`
	StartedAt       time.Time   `json:"startedAt"`
	TotalBackends   int         `json:"totalBackends"`
	HealthyBackends int         `json:"healthyBackends"`
}

type PodInfo struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

type Token struct {
	Token string `json:"access_token"`
}

type ResourceQuota struct {
	Namespace string                  `json:"namespace"`
	Data      v12.ResourceQuotaStatus `json:"data"`
}

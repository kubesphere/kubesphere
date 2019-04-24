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
	corev1 "k8s.io/api/core/v1"
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
	Username        string            `json:"username"`
	Email           string            `json:"email"`
	Lang            string            `json:"lang,omitempty"`
	Description     string            `json:"description"`
	CreateTime      time.Time         `json:"create_time"`
	Groups          []string          `json:"groups,omitempty"`
	Password        string            `json:"password,omitempty"`
	CurrentPassword string            `json:"current_password,omitempty"`
	AvatarUrl       string            `json:"avatar_url"`
	LastLoginTime   string            `json:"last_login_time"`
	Status          int               `json:"status"`
	ClusterRole     string            `json:"cluster_role"`
	Roles           map[string]string `json:"roles,omitempty"`
	Role            string            `json:"role,omitempty"`
	RoleBinding     string            `json:"role_binding,omitempty"`
	RoleBindTime    *time.Time        `json:"role_bind_time,omitempty"`
	WorkspaceRole   string            `json:"workspace_role,omitempty"`
}

type Group struct {
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Gid         string   `json:"gid"`
	Members     []string `json:"members"`
	Logo        string   `json:"logo"`
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
	Namespace string                     `json:"namespace"`
	Data      corev1.ResourceQuotaStatus `json:"data"`
}

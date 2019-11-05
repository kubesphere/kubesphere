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
	Items      []interface{} `json:"items" description:"paging data"`
	TotalCount int           `json:"total_count" description:"total count"`
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
	Name    string   `json:"name" description:"rule name"`
	Actions []string `json:"actions" description:"actions"`
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

type ComponentStatus struct {
	Name            string      `json:"name" description:"component name"`
	Namespace       string      `json:"namespace" description:"the name of the namespace"`
	SelfLink        string      `json:"selfLink" description:"self link"`
	Label           interface{} `json:"label" description:"labels"`
	StartedAt       time.Time   `json:"startedAt" description:"started time"`
	TotalBackends   int         `json:"totalBackends" description:"the total replicas of each backend system component"`
	HealthyBackends int         `json:"healthyBackends" description:"the number of healthy backend components"`
}
type NodeStatus struct {
	TotalNodes   int `json:"totalNodes" description:"total number of nodes"`
	HealthyNodes int `json:"healthyNodes" description:"the number of healthy nodes"`
}

type HealthStatus struct {
	KubeSphereComponents []ComponentStatus `json:"kubesphereStatus" description:"kubesphere components status"`
	NodeStatus           NodeStatus        `json:"nodeStatus" description:"nodes status"`
}

type PodInfo struct {
	Namespace string `json:"namespace" description:"namespace"`
	Pod       string `json:"pod" description:"pod name"`
	Container string `json:"container" description:"container name"`
}

type AuthGrantResponse struct {
	TokenType    string  `json:"token_type,omitempty"`
	Token        string  `json:"access_token" description:"access token"`
	ExpiresIn    float64 `json:"expires_in,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
}

type ResourceQuota struct {
	Namespace string                     `json:"namespace" description:"namespace"`
	Data      corev1.ResourceQuotaStatus `json:"data" description:"resource quota status"`
}

package iam

import (
	"k8s.io/api/rbac/v1"
)

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
	Username       string                  `json:"username"`
	Groups         []string                `json:"groups"`
	Password       string                  `json:"password,omitempty"`
	AvatarUrl      string                  `json:"avatar_url"`
	Description    string                  `json:"description"`
	Email          string                  `json:"email"`
	LastLoginTime  string                  `json:"last_login_time"`
	Status         int                     `json:"status"`
	ClusterRole    string                  `json:"cluster_role"`
	ClusterRules   []SimpleRule            `json:"cluster_rules,omitempty"`
	Roles          map[string]string       `json:"roles,omitempty"`
	Rules          map[string][]SimpleRule `json:"rules,omitempty"`
	Role           string                  `json:"role,omitempty"`
	WorkspaceRoles map[string]string       `json:"workspace_roles,omitempty"`
	WorkspaceRole  string                  `json:"workspace_role,omitempty"`
	WorkspaceRules map[string][]SimpleRule `json:"workspace_rules,omitempty"`
}

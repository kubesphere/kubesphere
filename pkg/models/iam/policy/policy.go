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

package policy

import (
	"encoding/json"
	"io/ioutil"

	"kubesphere.io/kubesphere/pkg/models"

	"k8s.io/api/rbac/v1"
)

const (
	rulesConfigPath        = "/etc/kubesphere/rules/rules.json"
	clusterRulesConfigPath = "/etc/kubesphere/rules/clusterrules.json"
)

func init() {
	rulesConfig, err := ioutil.ReadFile(rulesConfigPath)

	if err == nil {
		config := &[]models.Rule{}
		json.Unmarshal(rulesConfig, config)
		if len(*config) > 0 {
			RoleRuleMapping = *config
		}
	}

	clusterRulesConfig, err := ioutil.ReadFile(clusterRulesConfigPath)

	if err == nil {
		config := &[]models.Rule{}
		json.Unmarshal(clusterRulesConfig, config)
		if len(*config) > 0 {
			ClusterRoleRuleMapping = *config
		}
	}
}

var (
	WorkspaceRoleRuleMapping = []models.Rule{
		{
			Name: "workspaces",
			Actions: []models.Action{

				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
						}, {
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/*"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"jenkins.kubesphere.io"},
							Resources: []string{"*"},
						}, {
							Verbs:     []string{"*"},
							APIGroups: []string{"devops.kubesphere.io"},
							Resources: []string{"*"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
						},
					},
				},
			},
		},

		{Name: "members",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/members"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/members"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"patch", "update"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/members"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/members"},
						},
					},
				},
			},
		},
		{
			Name: "devops",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/devops"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/devops"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/devops"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/devops"},
						},
					},
				},
			},
		},
		{
			Name: "projects",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/namespaces"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/namespaces"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/namespaces"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/namespaces"},
						},
					},
				},
			},
		},
		{
			Name: "organizations",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"workspaces/organizations"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"workspaces/organizations"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"workspaces/organizations"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"workspaces/organizations"},
						},
					},
				}},
		},
		{
			Name: "roles",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/roles"},
						},
					}},
			},
		},
	}

	ClusterRoleRuleMapping = []models.Rule{
		{Name: "workspaces",
			Actions: []models.Action{
				{
					Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"workspaces"},
							Resources:     []string{"monitoring/*"},
						},
						{
							Verbs:     []string{"list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"quota", "status", "monitoring", "persistentvolumeclaims"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"resources"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces", "workspaces/*"},
						},
						{
							Verbs:     []string{"get"},
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"", "apps", "extensions", "batch"},
							Resources: []string{"serviceaccounts", "limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumeclaims", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings", "roles"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"members"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"jenkins.kubesphere.io", "devops.kubesphere.io"},
							Resources: []string{"*"},
						},
					},
				},
				{
					Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces", "workspaces/*"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"", "apps", "extensions", "batch"},
							Resources: []string{"serviceaccounts", "limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumeclaims", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings", "roles"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"members"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"jenkins.kubesphere.io", "devops.kubesphere.io"},
							Resources: []string{"*"},
						},
					},
				},
			},
		},
		{
			Name: "monitoring",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"*"},
						APIGroups: []string{"kubesphere.io"},
						Resources: []string{"monitoring", "health", "monitoring/*"},
					}},
				},
			},
		},
		{
			Name: "accounts",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users", "users/*"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"account.kubesphere.io"},
							Resources:     []string{"clusterrules"},
							ResourceNames: []string{"mapping"},
						},
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterrolebindings"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create", "get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"account.kubesphere.io"},
							Resources:     []string{"clusterrules"},
							ResourceNames: []string{"mapping"},
						},
						{
							Verbs:     []string{"create", "delete", "deletecollection"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterrolebindings"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list", "update", "patch"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users"},
						},
						{
							Verbs:     []string{"create", "delete", "deletecollection"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterrolebindings"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"accounts"},
						},
					},
				},
			},
		}, {
			Name: "roles",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterroles"},
						},
						{
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"cluster-roles"},
							Resources:     []string{"resources"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"clusterroles/*"},
						},
					},
				},

				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterroles"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterroles"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterroles"},
						},
					},
				},
			},
		}, {
			Name: "storageclasses",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"storage.k8s.io"},
							Resources: []string{"storageclasses"},
						}, {
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"storage-classes"},
							Resources:     []string{"resources"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"storage/*"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"storage.k8s.io"},
							Resources: []string{"storageclasses"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"storage.k8s.io"},
							Resources: []string{"storageclasses"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"storage.k8s.io"},
							Resources: []string{"storageclasses"},
						},
					},
				},
			},
		}, {
			Name: "nodes",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"nodes", "events"},
						},
						{
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"nodes"},
							Resources:     []string{"resources", "monitoring", "monitoring/*"},
						}, {
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"pods"},
							Resources:     []string{"resources"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{""},
							Resources: []string{"nodes"},
						},
					},
				},
			},
		}, {
			Name: "repos",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos"},
						},
					},
				},
			},
		}, {
			Name: "apps",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"list"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos", "app_versions"},
						}, {
							Verbs:     []string{"get"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"app_version/*"},
						},
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"apps", "clusters"},
						},
					},
				},
			},
		}, {
			Name: "components",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"list", "get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"components", "components/*"},
						},
						{
							Verbs:     []string{"list", "get"},
							APIGroups: []string{""},
							Resources: []string{"pods"},
						},
					},
				},
			},
		}}

	RoleRuleMapping = []models.Rule{{
		Name: "projects",
		Actions: []models.Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"events"},
					},
				},
			},
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
		},
	},
		{
			Name: "members",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list", "create", "update", "patch"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings"},
						},
					},
				},
			},
		},
		{
			Name: "roles",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"roles"},
						},
						{
							Verbs:         []string{"get", "list"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"roles"},
							Resources:     []string{"resources"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"roles"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"patch", "update"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"roles"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"roles"},
						},
					},
				},
			},
		},
		{
			Name: "deployments",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments", "deployments/scale"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/*"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments"},
						},
					},
				},

				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments", "deployments/*"},
						},
					},
				},

				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments"},
						},
					},
				},
				{Name: "scale",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments/scale"},
						},
					},
				},
			},
		}, {
			Name: "statefulsets",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/*"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
					},
				},
				{Name: "scale",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"patch"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
					},
				},
			},
		}, {
			Name: "daemonsets",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/*"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
					},
				},
			},
		}, {
			Name: "pods",
			Actions: []models.Action{
				{Name: "terminal",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"pod/shell"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"*"},
							Resources: []string{"pods"},
						},
					},
				},
			},
		},
		{
			Name: "services",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"list", "get"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
					},
				},

				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
					},
				},
			},
		},
		{
			Name: "internet",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
			},
		},

		{
			Name: "routes",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
					},
				},
			},
		}, {
			Name: "volumes",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumeclaims"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumeclaims"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumeclaims"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumeclaims"},
						},
					},
				},
			},
		}, {
			Name: "applications",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"applications"},
							Resources:     []string{"resources"},
						},
						{
							Verbs:     []string{"list"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"repos", "app_versions"},
						}, {
							Verbs:     []string{"get"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"app_version/*"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"apps"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:         []string{"create"},
							APIGroups:     []string{"openpitrix.io"},
							ResourceNames: []string{"delete"},
							Resources:     []string{"clusters"},
						},
					},
				},
			},
		},
		{
			Name: "jobs",
			Actions: []models.Action{
				{Name: "view", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"view", "list"},
						APIGroups: []string{"batch"},
						Resources: []string{"jobs"},
					},
				}},
				{Name: "create", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{"batch"},
						Resources: []string{"jobs"},
					},
				}},
				{Name: "edit", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{"batch"},
						Resources: []string{"jobs"},
					},
				}},
				{Name: "delete", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{"batch"},
						Resources: []string{"jobs"},
					},
				}},
			},
		},
		{
			Name: "cronjobs",
			Actions: []models.Action{
				{Name: "view", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"view", "list"},
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs"},
					},
				}},
				{Name: "create", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs"},
					},
				}},
				{Name: "edit", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs"},
					},
				}},
				{Name: "delete", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs"},
					},
				}},
			},
		},
		{
			Name: "secrets",
			Actions: []models.Action{
				{Name: "view", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"view", "list"},
						APIGroups: []string{""},
						Resources: []string{"secrets"},
					},
				}},
				{Name: "create", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{""},
						Resources: []string{"secrets"},
					},
				}},
				{Name: "edit", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{""},
						Resources: []string{"secrets"},
					},
				}},
				{Name: "delete", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{""},
						Resources: []string{"secrets"},
					},
				}},
			},
		},
		{
			Name: "configmaps",
			Actions: []models.Action{
				{Name: "view", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"view", "list"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
				{Name: "create", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
				{Name: "edit", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
				{Name: "delete", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
			},
		},
	}
)

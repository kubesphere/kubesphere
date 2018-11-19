/*
 Copyright 2018 The KubeSphere Authors.

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

package iam

import (
	"encoding/json"
	"io/ioutil"

	"k8s.io/api/rbac/v1"
)

const (
	rulesConfigPath        = "/etc/kubesphere/rules/rules.json"
	clusterRulesConfigPath = "/etc/kubesphere/rules/clusterrules.json"
)

func init() {
	rulesConfig, err := ioutil.ReadFile(rulesConfigPath)

	if err == nil {
		config := &[]Rule{}
		json.Unmarshal(rulesConfig, config)
		if len(*config) > 0 {
			RoleRuleMapping = *config
		}
	}

	clusterRulesConfig, err := ioutil.ReadFile(clusterRulesConfigPath)

	if err == nil {
		config := &[]Rule{}
		json.Unmarshal(clusterRulesConfig, config)
		if len(*config) > 0 {
			ClusterRoleRuleMapping = *config
		}
	}
}

var (
	WorkspaceRoleRuleMapping = []Rule{
		{
			Name: "workspaces",
			Actions: []Action{

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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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

	ClusterRoleRuleMapping = []Rule{
		{Name: "workspaces",
			Actions: []Action{
				{
					Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
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
							Verbs:     []string{"get", "watch", "list", "create", "delete", "patch", "update"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces", "workspaces/namespaces", "workspaces/roles", "workspaces/devops", "workspaces/members"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"workspaces"},
							Resources:     []string{"monitoring/*"},
						},
					},
				},
			},
		},
		{
			Name: "accounts",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"accounts"},
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
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"accounts"},
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
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"accounts"},
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"clusterroles"},
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"storage.k8s.io"},
							Resources: []string{"storageclasses"},
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"nodes"},
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
				{Name: "cordon",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{""},
							Resources: []string{"nodes"},
						},
					}},
				{Name: "taint", Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{""},
						Resources: []string{"nodes"},
					},
				}},
			},
		}, {
			Name: "repos",
			Actions: []Action{
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"openpitrix.io"},
							Resources: []string{"apps", "repos"},
						},
					},
				},
			},
		}, {
			Name: "components",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"kubsphere.io"},
							Resources: []string{"components"},
						},
					},
				},
			},
		}}

	RoleRuleMapping = []Rule{{
		Name: "projects",
		Actions: []Action{
			//  limit range +  router
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch", "get"},
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
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
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
		},
	},
		{
			Name: "members",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list", "create", "delete"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"rolebindings"},
						},
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
			},
		},
		{
			Name: "roles",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list", "create", "delete", "patch", "update"},
							APIGroups: []string{"rbac.authorization.k8s.io"},
							Resources: []string{"roles"},
						},
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
			},
		},
		{
			Name: "deployments",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments", "deployments/scale"},
						},
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
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/log", "pods/status"},
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
							Resources: []string{"deployments", "deployments/rollback"},
						},
					},
				},

				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments"},
						},
					},
				},
				{Name: "scale",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create", "update", "patch", "delete"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"deployments/scale"},
						},
					},
				},
			},
		}, {
			Name: "statefulsets",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"apps"},
							Resources: []string{"statefulsets"},
						},
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
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/log", "pods/status"},
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
							Verbs:     []string{"delete", "deletecollection"},
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
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
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"pods", "pods/log", "pods/status"},
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
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"apps", "extensions"},
							Resources: []string{"daemonsets"},
						},
					},
				},
			},
		}, {
			Name: "pods",
			Actions: []Action{
				{Name: "terminal",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"terminal"},
						},
					},
				},
			},
		},
		{
			Name: "services",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"list"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
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
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{""},
							Resources: []string{"services"},
						},
					},
				},
			},
		}, {
			Name: "routes",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
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
							Verbs:     []string{"delete", "deletecollection"},
							APIGroups: []string{"extensions"},
							Resources: []string{"ingresses"},
						},
					},
				},
			},
		}, {
			Name: "volumes",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumes"},
						},
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
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumes"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumes"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "watch", "list"},
							APIGroups: []string{""},
							Resources: []string{"persistentvolumes"},
						},
					},
				},
			},
		}, {
			Name: "applications",
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"applications"},
						},
					},
				},
			},
		},
		{
			Name: "jobs",
			Actions: []Action{
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{"batch"},
						Resources: []string{"jobs"},
					},
				}},
			},
		},
		{
			Name: "cronjobs",
			Actions: []Action{
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{"batch"},
						Resources: []string{"cronjobs"},
					},
				}},
			},
		},
		{
			Name: "secrets",
			Actions: []Action{
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{""},
						Resources: []string{"secrets"},
					},
				}},
			},
		},
		{
			Name: "configmaps",
			Actions: []Action{
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
			},
		},
	}
)

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
	ClusterRoleRuleMapping = []models.Rule{
		{Name: "workspaces",
			Actions: []models.Action{
				{
					Name: "manage",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"*"},
							Resources: []string{"workspaces", "workspaces/*"},
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
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"monitoring.kubesphere.io"},
						Resources: []string{"*"},
					}, {
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"resources.kubesphere.io"},
						Resources: []string{"health"},
					}},
				},
			},
		},
		{
			Name: "alerting",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"create"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"delete"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
				},
			},
		},
		{
			Name: "logging",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"logging.kubesphere.io"},
						Resources: []string{"*"},
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
							APIGroups: []string{"iam.kubesphere.io"},
							Resources: []string{"users", "users/*"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"iam.kubesphere.io"},
							Resources:     []string{"rulesmapping"},
							ResourceNames: []string{"clusterroles"},
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
							APIGroups: []string{"iam.kubesphere.io"},
							Resources: []string{"users"},
						},
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"iam.kubesphere.io"},
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
							APIGroups: []string{"iam.kubesphere.io"},
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
							APIGroups: []string{"iam.kubesphere.io"},
							Resources: []string{"users"},
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
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"iam.kubesphere.io"},
							Resources: []string{"clusterroles", "clusterroles/*"},
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
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"storageclasses", "storageclasses/*"},
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
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"nodes", "nodes/*"},
						}, {
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"monitoring.kubesphere.io"},
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
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"components", "components/*"},
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
						APIGroups: []string{"*"},
						Resources: []string{"namespaces"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{"*"},
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
			Name: "monitoring",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"monitoring.kubesphere.io"},
						Resources: []string{"*"},
					}, {
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"resources.kubesphere.io"},
						Resources: []string{"health"},
					}},
				},
			},
		},

		{
			Name: "alerting",
			Actions: []models.Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"create"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"delete"},
						APIGroups: []string{"alerting.kubesphere.io"},
						Resources: []string{"*"},
					}},
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
							APIGroups: []string{"rbac.authorization.k8s.io", "resources.kubesphere.io"},
							Resources: []string{"rolebindings"},
						},
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"iam.kubesphere.io"},
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
							APIGroups: []string{"rbac.authorization.k8s.io", "resources.kubesphere.io"},
							Resources: []string{"roles"},
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
							APIGroups: []string{"apps", "extensions", "resources.kubesphere.io"},
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
							APIGroups: []string{"apps", "resources.kubesphere.io"},
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
							APIGroups: []string{"apps", "extensions", "resources.kubesphere.io"},
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
							APIGroups: []string{"terminal.kubesphere.io"},
							Resources: []string{"pods"},
						},
					},
				},
				{Name: "view",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"pods"},
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
							APIGroups: []string{"", "resources.kubesphere.io"},
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
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "create",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "edit",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"update", "patch"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"router"},
						},
					},
				},
				{Name: "delete",
					Rules: []v1.PolicyRule{
						{
							Verbs:     []string{"delete"},
							APIGroups: []string{"resources.kubesphere.io"},
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
							APIGroups: []string{"extensions", "resources.kubesphere.io"},
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
							APIGroups: []string{"", "resources.kubesphere.io"},
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
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"resources.kubesphere.io"},
							Resources: []string{"applications"},
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
						APIGroups: []string{"batch", "resources.kubesphere.io"},
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
						APIGroups: []string{"batch", "resources.kubesphere.io"},
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
						APIGroups: []string{"", "resources.kubesphere.io"},
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
						APIGroups: []string{"", "resources.kubesphere.io"},
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

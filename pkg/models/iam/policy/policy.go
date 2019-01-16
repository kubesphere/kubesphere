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

	. "kubesphere.io/kubesphere/pkg/models"

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
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/quota/namespaces/xxx
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
						}, {
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces/*"},
						},
						////patch api/v1/namespaces/xxx
						//{
						//	Verbs:     []string{"*"},
						//	APIGroups: []string{""},
						//	Resources: []string{"namespaces"},
						//},
						//post api/v1/namespaces/xxx/resourcequotas/xxx
						//post api/v1/namespaces/xxx/resourcequotas
						//post api/v1/namespaces/xxx/serviceaccounts
						//post api/v1/namespaces/xxx/configmaps
						//post api/v1/namespaces/xxx/secrets
						//post apis/apps/v1/namespaces/xxx/deployments
						//post api/v1/namespaces/xxx/limitranges
						//{
						//	Verbs:     []string{"*"},
						//	APIGroups: []string{"", "apps", "extensions", "batch"},
						//	Resources: []string{"limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumeclaims", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
						//},
						//* apis/rbac.authorization.k8s.io/v1/namespaces/xxx/rolebindings
						//* apis/rbac.authorization.k8s.io/v1/namespaces/xxx/roles
						//{
						//	Verbs:     []string{"*"},
						//	APIGroups: []string{"rbac.authorization.k8s.io"},
						//	Resources: []string{"rolebindings", "roles"},
						//},
						//post apis/kubesphere.io/v1alpha1/namespaces/xxx/router
						//{
						//	Verbs:     []string{"create"},
						//	APIGroups: []string{"kubesphere.io"},
						//	Resources: []string{"router"},
						//},
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

	// 集群权限规则表
	ClusterRoleRuleMapping = []Rule{
		// 工作空间相关规则
		{Name: "workspaces",
			Actions: []Action{
				// 查看集群中所有的工作空间
				{
					Name: "view",
					Rules: []v1.PolicyRule{
						//get apis/account.kubesphere.io/v1alpha1/users
						//get apis/account.kubesphere.io/v1alpha1/namespaces/xxx/users
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"account.kubesphere.io"},
							Resources: []string{"users"},
						},
						//get apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample/_statistics
						//get apis/kubesphere.io/v1alpha1/monitoring/workspaces/sample
						{
							Verbs:         []string{"get"},
							APIGroups:     []string{"kubesphere.io"},
							ResourceNames: []string{"workspaces"},
							Resources:     []string{"monitoring/*"},
						},
						//get apis/kubesphere.io/v1alpha1/quota/namespaces/xxx
						//get apis/kubesphere.io/v1alpha1/status/namespaces/xxx
						//{
						//	Verbs:         []string{"get"},
						//	APIGroups:     []string{"kubesphere.io"},
						//	ResourceNames: []string{"namespaces"},
						//	Resources:     []string{"quota/*", "status/*"},
						//},
						{
							Verbs:     []string{"list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"quota", "status", "monitoring", "persistentvolumeclaims"},
						},
						//get apis/kubesphere.io/v1alpha1/monitoring/namespaces/xxx
						//{
						//	Verbs:         []string{"get"},
						//	APIGroups:     []string{"kubesphere.io"},
						//	ResourceNames: []string{"namespaces"},
						//	Resources:     []string{"monitoring/*"},
						//},
						//get apis/kubesphere.io/v1alpha1/resources/applications
						//get apis/kubesphere.io/v1alpha1/resources/deployments
						//get apis/kubesphere.io/v1alpha1/resources/statefulsets
						//get apis/kubesphere.io/v1alpha1/resources/daemonsets
						//get apis/kubesphere.io/v1alpha1/resources/jobs
						//get apis/kubesphere.io/v1alpha1/resources/cronjobs
						//get apis/kubesphere.io/v1alpha1/resources/persistent-volume-claims
						//get apis/kubesphere.io/v1alpha1/resources/services
						//get apis/kubesphere.io/v1alpha1/resources/ingresses
						//get apis/kubesphere.io/v1alpha1/resources/secrets
						//get apis/kubesphere.io/v1alpha1/resources/configmaps
						//get apis/kubesphere.io/v1alpha1/resources/roles
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"resources"},
						},
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/quota/namespaces/xxx
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces", "workspaces/*"},
						},
						//get api/v1/namespaces/xxx
						{
							Verbs:     []string{"get"},
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
						},
						//post api/v1/namespaces/xxx/resourcequotas/xxx
						//post api/v1/namespaces/xxx/resourcequotas
						//post api/v1/namespaces/xxx/serviceaccounts
						//post api/v1/namespaces/xxx/configmaps
						//post api/v1/namespaces/xxx/secrets
						//post apis/apps/v1/namespaces/xxx/deployments
						//post api/v1/namespaces/xxx/limitranges
						{
							Verbs:     []string{"get", "list"},
							APIGroups: []string{"", "apps", "extensions", "batch"},
							Resources: []string{"serviceaccounts", "limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumeclaims", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
						},
						//get apis/rbac.authorization.k8s.io/v1/namespaces/xxx/rolebindings
						//get apis/rbac.authorization.k8s.io/v1/namespaces/xxx/roles
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
						//post apis/kubesphere.io/v1alpha1/namespaces/xxx/router
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
				// 在集群中创建工作空间
				{
					Name: "create",
					Rules: []v1.PolicyRule{
						//post apis/kubesphere.io/v1alpha1/workspaces
						{
							Verbs:     []string{"create"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces"},
						},
					},
				},
				// 管理集群中所有的工作空间
				{Name: "edit",
					Rules: []v1.PolicyRule{
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/namespaces
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/devops
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/workspaces/sample/members
						//post apis/kubesphere.io/v1alpha1/quota/namespaces/xxx
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"kubesphere.io"},
							Resources: []string{"workspaces", "workspaces/*"},
						},

						//patch api/v1/namespaces/xxx
						{
							Verbs:     []string{"*"},
							APIGroups: []string{""},
							Resources: []string{"namespaces"},
						},
						//post api/v1/namespaces/xxx/resourcequotas/xxx
						//post api/v1/namespaces/xxx/resourcequotas
						//post api/v1/namespaces/xxx/serviceaccounts
						//post api/v1/namespaces/xxx/configmaps
						//post api/v1/namespaces/xxx/secrets
						//post apis/apps/v1/namespaces/xxx/deployments
						//post api/v1/namespaces/xxx/limitranges
						{
							Verbs:     []string{"*"},
							APIGroups: []string{"", "apps", "extensions", "batch"},
							Resources: []string{"serviceaccounts", "limitranges", "deployments", "configmaps", "secrets", "jobs", "cronjobs", "persistentvolumeclaims", "statefulsets", "daemonsets", "ingresses", "services", "pods/*", "pods", "events", "deployments/scale"},
						},
						//* apis/rbac.authorization.k8s.io/v1/namespaces/xxx/rolebindings
						//* apis/rbac.authorization.k8s.io/v1/namespaces/xxx/roles
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
						//post apis/kubesphere.io/v1alpha1/namespaces/xxx/router
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
			Actions: []Action{
				{Name: "view",
					Rules: []v1.PolicyRule{{
						Verbs:     []string{"*"},
						APIGroups: []string{"kubesphere.io"},
						Resources: []string{"monitoring", "health", "monitoring/*"},
					}},
				},
			},
		},
		//{
		//	Name: "kubectl",
		//	Actions: []models.Action{
		//		{Name: "view",
		//			Rules: []v1.PolicyRule{{
		//				Verbs:     []string{"get"},
		//				APIGroups: []string{"kubesphere.io"},
		//				Resources: []string{"pod/shell"},
		//			}},
		//		},
		//	},
		//},
		//{
		//	Name: "projects",
		//	Actions: []models.Action{
		//		{Name: "view",
		//			Rules: []v1.PolicyRule{
		//				{
		//					Verbs:     []string{"get", "watch", "list"},
		//					APIGroups: []string{""},
		//					Resources: []string{"namespaces"},
		//				},
		//			},
		//		},
		//		{Name: "create",
		//			Rules: []v1.PolicyRule{
		//				{
		//					Verbs:     []string{"create"},
		//					APIGroups: []string{""},
		//					Resources: []string{"namespaces"},
		//				},
		//			},
		//		},
		//		{Name: "edit",
		//			Rules: []v1.PolicyRule{
		//				{
		//					Verbs:     []string{"update", "patch"},
		//					APIGroups: []string{""},
		//					Resources: []string{"namespaces"},
		//				},
		//			},
		//		},
		//		{Name: "delete",
		//			Rules: []v1.PolicyRule{
		//				{
		//					Verbs:     []string{"delete", "deletecollection"},
		//					APIGroups: []string{""},
		//					Resources: []string{"namespaces"},
		//				},
		//			},
		//		},
		//		{Name: "members",
		//			Rules: []v1.PolicyRule{
		//				{
		//					Verbs:     []string{"get", "watch", "list", "create", "delete", "patch", "update"},
		//					APIGroups: []string{"rbac.authorization.k8s.io"},
		//					Resources: []string{"rolebindings", "roles"},
		//				},
		//			},
		//		},
		//	},
		//},
		{
			Name: "accounts",
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
				//{Name: "cordon",
				//	Rules: []v1.PolicyRule{
				//		{
				//			Verbs:     []string{"update", "patch"},
				//			APIGroups: []string{""},
				//			Resources: []string{"nodes"},
				//		},
				//	},},
				//{Name: "taint", Rules: []v1.PolicyRule{
				//	{
				//		Verbs:     []string{"update", "patch"},
				//		APIGroups: []string{""},
				//		Resources: []string{"nodes"},
				//	},
				//},},
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
			//应用模板
			Name: "apps",
			Actions: []Action{
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
			//服务组件 apis/kubesphere.io/v1alpha1/components
			Name: "components",
			Actions: []Action{
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

	RoleRuleMapping = []Rule{{
		Name: "projects",
		Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
			Actions: []Action{
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
						Verbs:     []string{"delete"},
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
						Verbs:     []string{"delete"},
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
						Verbs:     []string{"delete"},
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
						Verbs:     []string{"delete"},
						APIGroups: []string{""},
						Resources: []string{"configmaps"},
					},
				}},
			},
		},
	}
)

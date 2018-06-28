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

type Action struct {
	Name  string          `json:"name"`
	Rules []v1.PolicyRule `json:"rules"`
}

type Rule struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

func init() {
	rulesConfig, err := ioutil.ReadFile(rulesConfigPath)
	if err == nil {
		config := &[]Rule{}
		json.Unmarshal(rulesConfig, config)
		if len(*config) > 0 {
			RoleRuleGroup = *config
		}
	}

	clusterRulesConfig, err := ioutil.ReadFile(clusterRulesConfigPath)

	if err == nil {
		config := &[]Rule{}
		json.Unmarshal(clusterRulesConfig, config)
		if len(*config) > 0 {
			ClusterRoleRuleGroup = *config
		}
	}
}

var (
	ClusterRoleRuleGroup = []Rule{projects, users, roles, images,
		volumes, storageclasses, nodes, appCatalog, apps, components, deployments, statefulsets, daemonsets, pods, services, routes}

	RoleRuleGroup = []Rule{project, deployments, statefulsets, daemonsets, pods,
		services, routes, volumes}

	components = Rule{
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
	}

	projects = Rule{
		Name: "projects",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
			{Name: "create",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
			{Name: "members",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
					},
				},
			},
			{Name: "member_roles",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list", "create", "delete", "patch", "update"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles"},
					},
				},
			},
		},
	}

	project = Rule{
		Name: "projects",
		Actions: []Action{
			{Name: "members",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
					},
				},
			},
			{Name: "member_roles",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list", "create", "delete", "patch", "update"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles"},
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
	}
	users = Rule{
		Name: "users",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{"kubesphere.io"},
						Resources: []string{"users"},
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
						Resources: []string{"users"},
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
						Resources: []string{"users"},
					},
				},
			},
		},
	}

	roles = Rule{
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
	}

	nodes = Rule{
		Name: "nodes",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{""},
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
			{Name: "drain",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"*"},
						APIGroups: []string{"kubesphere.io"},
						Resources: []string{"nodes"},
					},
				},
			},
		},
	}

	volumes = Rule{
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
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
	}

	storageclasses = Rule{
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
	}

	images = Rule{
		Name: "images",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{""},
						Resources: []string{
							"secrets",
						},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
			{Name: "create",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{""},
						Resources: []string{
							"secrets",
						},
					},
				},
			},
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{""},
						Resources: []string{
							"secrets",
						},
					},
				},
			},
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{""},
						Resources: []string{
							"secrets",
						},
					},
				},
			},
		},
	}

	appCatalog = Rule{
		Name: "app_catalog",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"appcatalog"},
					},
				},
			},
			{Name: "create",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"appcatalog"},
					},
				},
			},
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"appcatalog"},
					},
				},
			},
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"appcatalog"},
					},
				},
			},
		},
	}

	apps = Rule{
		Name: "apps",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"apps"},
					},
				},
			},
		},
	}

	statefulsets = Rule{
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
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
	}

	daemonsets = Rule{
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
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
	}

	services = Rule{
		Name: "services",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
						APIGroups: []string{""},
						Resources: []string{"services"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
	}

	routes = Rule{
		Name: "routes",
		Actions: []Action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "watch", "list"},
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
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{"extensions"},
						Resources: []string{"ingresses"},
					},
				},
			},
		},
	}

	deployments = Rule{
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
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
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
	}

	pods = Rule{
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
	}
)

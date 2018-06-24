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

	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
)

const (
	rulesConfigPath        = "/etc/kubesphere/rules/rules.json"
	clusterRulesConfigPath = "/etc/kubesphere/rules/clusterrules.json"
)

type roleList struct {
	ClusterRoles []v1.ClusterRole `json:"clusterRoles" protobuf:"bytes,2,rep,name=clusterRoles"`
	Roles        []v1.Role        `json:"roles" protobuf:"bytes,2,rep,name=roles"`
}

type action struct {
	Name  string          `json:"name"`
	Rules []v1.PolicyRule `json:"rules"`
}

type rule struct {
	Name    string   `json:"name"`
	Actions []action `json:"actions"`
}

type userRuleList struct {
	ClusterRules []rule            `json:"clusterRules"`
	Rules        map[string][]rule `json:"rules"`
}

func init() {
	rulesConfig, err := ioutil.ReadFile(rulesConfigPath)
	if err == nil {
		config := &[]rule{}
		json.Unmarshal(rulesConfig, config)
		if len(*config) > 0 {
			roleRuleGroup = *config
			glog.Info("rules config load success")
		}
	}

	clusterRulesConfig, err := ioutil.ReadFile(clusterRulesConfigPath)

	if err == nil {
		config := &[]rule{}
		json.Unmarshal(clusterRulesConfig, config)
		if len(*config) > 0 {
			clusterRoleRuleGroup = *config
			glog.Info("cluster rules config load success")
		}
	}
}

var (
	clusterRoleRuleGroup = []rule{projects, users, roles, images,
		volumes, storageclasses, nodes, appCatalog, apps, components, deployments, statefulsets, daemonsets, pods, services, routes}

	roleRuleGroup = []rule{project, deployments, statefulsets, daemonsets, pods,
		services, routes, volumes}

	components = rule{
		Name: "components",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"kubsphere.io"},
						Resources: []string{"components"},
					},
				},
			},
		},
	}

	projects = rule{
		Name: "projects",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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
						Verbs:     []string{"get", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
					},
				},
			},
			{Name: "member_roles",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "create", "delete", "patch", "update"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles"},
					},
				},
			},
		},
	}

	project = rule{
		Name: "projects",
		Actions: []action{
			{Name: "members",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
					},
				},
			},
			{Name: "member_roles",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "create", "delete", "patch", "update"},
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
	users = rule{
		Name: "users",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"kubesphere.io"},
						Resources: []string{"users"},
					},
					{
						Verbs:     []string{"get", "list"},
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

	roles = rule{
		Name: "roles",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	nodes = rule{
		Name: "nodes",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	volumes = rule{
		Name: "volumes",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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
						Verbs:     []string{"get", "list"},
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
						Verbs:     []string{"get", "list"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumes"},
					},
				},
			},
		},
	}

	storageclasses = rule{
		Name: "storageclasses",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	images = rule{
		Name: "images",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	appCatalog = rule{
		Name: "app_catalog",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	apps = rule{
		Name: "apps",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"openpitrix.io"},
						Resources: []string{"apps"},
					},
				},
			},
		},
	}

	statefulsets = rule{
		Name: "statefulsets",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps"},
						Resources: []string{"statefulsets"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
					{
						Verbs:     []string{"get", "list"},
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

	daemonsets = rule{
		Name: "daemonsets",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps", "extensions"},
						Resources: []string{"daemonsets"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
					{
						Verbs:     []string{"get", "list"},
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

	services = rule{
		Name: "services",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
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

	routes = rule{
		Name: "routes",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"extensions"},
						Resources: []string{"ingresses"},
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

	deployments = rule{
		Name: "deployments",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps", "extensions"},
						Resources: []string{"deployments", "deployments/scale"},
					},
					{
						Verbs:     []string{"list"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
					{
						Verbs:     []string{"get", "list"},
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

	pods = rule{
		Name: "pods",
		Actions: []action{
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

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
	"k8s.io/api/rbac/v1"
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

// TODO stored in etcd, allow updates
var (
	clusterRoleRuleGroup = []rule{projectsManagement, userManagement, roleManagement, registryManagement,
		volumeManagement, storageclassManagement, nodeManagement, appCatalogManagement, appManagement}

	roleRuleGroup = []rule{deploymentManagement, projectManagement, statefulsetManagement, daemonsetManagement,
		serviceManagement, routeManagement, pvcManagement}

	projectsManagement = rule{
		Name: "projectsManagement",
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
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{""},
						Resources: []string{"namespaces"},
					},
				},
			},
		},
	}

	userManagement = rule{
		Name: "userManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"iam.kubesphere.io"},
						Resources: []string{"users"},
					},
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings", "clusterrolebindings"},
					},
				},
			},
			{Name: "create",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{"iam.kubesphere.io"},
						Resources: []string{"users"},
					},
				},
			},
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{"iam.kubesphere.io"},
						Resources: []string{"users"},
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
	}

	roleManagement = rule{
		Name: "roleManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles", "clusterroles"},
					},
				},
			},

			{Name: "create",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles", "clusterroles"},
					},
				},
			},
			{Name: "edit",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"update", "patch"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles", "clusterroles"},
					},
				},
			},
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete", "deletecollection"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles", "clusterroles"},
					},
				},
			},
			{Name: "roleBinding",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"create", "delete", "deletecollection"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings", "clusterrolebindings"},
					},
				},
			},
		},
	}

	nodeManagement = rule{
		Name: "nodeManagement",
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
		},
	}

	volumeManagement = rule{
		Name: "volumeManagement",
		Actions: []action{
			{Name: "view",
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

	storageclassManagement = rule{
		Name: "storageclassManagement",
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
		},
	}

	registryManagement = rule{
		Name: "registryManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"extend.kubesphere.io"},
						Resources: []string{
							"registries",
						},
					},
				},
			},
		},
	}

	appCatalogManagement = rule{
		Name: "appCatalogManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"extend.kubesphere.io"},
						Resources: []string{"appcatalog"},
					},
				},
			},
		},
	}

	appManagement = rule{
		Name: "appManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"extend.kubesphere.io"},
						Resources: []string{"apps"},
					},
				},
			},
		},
	}

	statefulsetManagement = rule{
		Name: "statefulsetManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps"},
						Resources: []string{"statefulsets"},
					},
				},
			},
		},
	}

	daemonsetManagement = rule{
		Name: "daemonsetManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps", "extensions"},
						Resources: []string{"daemonsets"},
					},
				},
			},
		},
	}

	serviceManagement = rule{
		Name: "serviceManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{""},
						Resources: []string{"services"},
					},
				},
			},
		},
	}

	routeManagement = rule{
		Name: "routeManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"extensions"},
						Resources: []string{"ingresses"},
					},
				},
			},
		},
	}
	pvcManagement = rule{
		Name: "pvcManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{""},
						Resources: []string{"persistentvolumeclaims"},
					},
				},
			},
		},
	}

	deploymentManagement = rule{
		Name: "deploymentManagement",
		Actions: []action{
			{Name: "view",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list"},
						APIGroups: []string{"apps", "extensions"},
						Resources: []string{
							"deployments",
							"deployments/rollback",
							"deployments/scale",
						},
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
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete", "deletecollection"},
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
	projectManagement = rule{
		Name: "projectManagement",
		Actions: []action{
			{Name: "memberManagement",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"rolebindings"},
					},
				},
			},
			{Name: "memberRoleManagement",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"get", "list", "create", "delete"},
						APIGroups: []string{"rbac.authorization.k8s.io"},
						Resources: []string{"roles"},
					},
				},
			},
			{Name: "delete",
				Rules: []v1.PolicyRule{
					{
						Verbs:     []string{"delete"},
						APIGroups: []string{"extend.kubesphere.io"},
						Resources: []string{"namespace"},
					},
				},
			},
		},
	}
)

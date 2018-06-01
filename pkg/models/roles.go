package models

import (
	"k8s.io/api/rbac/v1"
	"kubesphere.io/kubesphere/pkg/client"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ClusterRoleKind = "ClusterRole"

func GetRole(namespace string, name string) (*v1.Role, error) {
	k8s := client.NewK8sClient()
	role, err := k8s.RbacV1().Roles(namespace).Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func GetClusterRoleBindings(name string) ([]v1.ClusterRoleBinding, error) {
	k8s := client.NewK8sClient()
	roleBindingList, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	items := make([]v1.ClusterRoleBinding, 0)

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.Name == name {
			items = append(items, roleBinding)
		}
	}

	return roleBindingList.Items, nil
}

func GetRoleBindings(namespace string, name string) ([]v1.RoleBinding, error) {
	k8s := client.NewK8sClient()

	roleBindingList, err := k8s.RbacV1().RoleBindings(namespace).List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	items := make([]v1.RoleBinding, 0)

	for _, roleBinding := range roleBindingList.Items {
		if roleBinding.RoleRef.Name == name {
			items = append(items, roleBinding)
		}
	}

	return roleBindingList.Items, nil
}

func GetClusterRole(name string) (*v1.ClusterRole, error) {
	k8s := client.NewK8sClient()
	role, err := k8s.RbacV1().ClusterRoles().Get(name, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return role, nil
}

func GetRoles(username string) ([]v1.Role, error) {
	k8s := client.NewK8sClient()

	roleBindings, err := k8s.RbacV1().RoleBindings("").List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}
	roles := make([]v1.Role, 0)

	for _, roleBinding := range roleBindings.Items {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {

					clusterRole, err := k8s.RbacV1().ClusterRoles().Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})

					if err != nil {
						return nil, err
					}

					var role = v1.Role(*clusterRole)
					role.Namespace = roleBinding.Namespace

					roles = append(roles, role)

				} else {
					rule, err := k8s.RbacV1().Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})

					if err != nil {
						return nil, err
					}

					roles = append(roles, *rule)
				}
			}
		}
	}

	return roles, nil
}

func GetClusterRoles(username string) ([]v1.ClusterRole, error) {
	k8s := client.NewK8sClient()

	clusterRoleBindings, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	if err != nil {
		return nil, err
	}

	roles := make([]v1.ClusterRole, 0)

	for _, roleBinding := range clusterRoleBindings.Items {
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				if roleBinding.RoleRef.Kind == ClusterRoleKind {

					rule, err := k8s.RbacV1().ClusterRoles().Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})

					if err != nil {
						return nil, err
					}

					roles = append(roles, *rule)
				}
			}
		}
	}

	return roles, nil
}

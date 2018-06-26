package models

import (
	"github.com/golang/glog"
	"k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kubesphere.io/kubesphere/pkg/client"
)

const ClusterRoleKind = "ClusterRole"

func DeleteRoleBindings(username string) error {
	k8s := client.NewK8sClient()

	roleBindings, err := k8s.RbacV1().RoleBindings("").List(meta_v1.ListOptions{})

	if err != nil {
		return err
	}

	for _, roleBinding := range roleBindings.Items {

		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)

		if length2 == 0 {
			k8s.RbacV1().RoleBindings(roleBinding.Namespace).Delete(roleBinding.Name, &meta_v1.DeleteOptions{})
		} else if length2 < length1 {
			k8s.RbacV1().RoleBindings(roleBinding.Namespace).Update(&roleBinding)
		}
	}

	clusterRoleBindingList, err := k8s.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})

	for _, roleBinding := range clusterRoleBindingList.Items {
		length1 := len(roleBinding.Subjects)

		for index, subject := range roleBinding.Subjects {
			if subject.Kind == v1.UserKind && subject.Name == username {
				roleBinding.Subjects = append(roleBinding.Subjects[:index], roleBinding.Subjects[index+1:]...)
				index--
			}
		}

		length2 := len(roleBinding.Subjects)
		if length2 == 0 {
			k8s.RbacV1().ClusterRoleBindings().Delete(roleBinding.Name, &meta_v1.DeleteOptions{})
		} else if length2 < length1 {
			k8s.RbacV1().ClusterRoleBindings().Update(&roleBinding)
		}
	}

	return nil
}

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

	return items, nil
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

	return items, nil
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
					if err == nil {
						var role = v1.Role{TypeMeta: (*clusterRole).TypeMeta, ObjectMeta: (*clusterRole).ObjectMeta, Rules: (*clusterRole).Rules}
						role.Namespace = roleBinding.Namespace
						roles = append(roles, role)
						break
					} else if apierrors.IsNotFound(err) {
						glog.Infoln(err.Error())
						break
					} else {
						return nil, err
					}

				} else {
					if subject.Kind == v1.UserKind && subject.Name == username {
						rule, err := k8s.RbacV1().Roles(roleBinding.Namespace).Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})
						if err == nil {
							roles = append(roles, *rule)
							break
						} else if apierrors.IsNotFound(err) {
							glog.Infoln(err.Error())
							break
						} else {
							return nil, err
						}

					}

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
					role, err := k8s.RbacV1().ClusterRoles().Get(roleBinding.RoleRef.Name, meta_v1.GetOptions{})
					if err == nil {
						role.Annotations["rbac.authorization.k8s.io/clusterrolebinding"] = roleBinding.Name
						roles = append(roles, *role)
						break
					} else if apierrors.IsNotFound(err) {
						glog.Infoln(err.Error())
						break
					} else {
						return nil, err
					}
				}
			}
		}
	}

	return roles, nil
}

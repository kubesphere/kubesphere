/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package user

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1beta1 "kubesphere.io/api/tenant/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
)

type usersGetter struct {
	cache runtimeclient.Reader
}

func New(cache runtimeclient.Reader) v1alpha3.Interface {
	return &usersGetter{cache: cache}
}

func (d *usersGetter) Get(_, name string) (runtime.Object, error) {
	user := &iamv1beta1.User{}
	return user, d.cache.Get(context.Background(), types.NamespacedName{Name: name}, user)
}

func (d *usersGetter) List(_ string, query *query.Query) (*api.ListResult, error) {
	var users []*iamv1beta1.User
	var err error

	if namespace := query.Filters[iamv1beta1.ScopeNamespace]; namespace != "" {
		role := query.Filters[iamv1beta1.ResourcesSingularRole]
		users, err = d.listAllUsersInNamespace(string(namespace), string(role))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.ScopeNamespace)
		delete(query.Filters, iamv1beta1.ResourcesSingularRole)
	} else if workspace := query.Filters[iamv1beta1.ScopeWorkspace]; workspace != "" {
		workspaceRole := query.Filters[iamv1beta1.ResourcesSingularWorkspaceRole]
		users, err = d.listAllUsersInWorkspace(string(workspace), string(workspaceRole))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.ScopeWorkspace)
		delete(query.Filters, iamv1beta1.ResourcesSingularWorkspaceRole)
	} else if cluster := query.Filters[iamv1beta1.ScopeCluster]; cluster == "true" {
		clusterRole := query.Filters[iamv1beta1.ResourcesSingularClusterRole]
		users, err = d.listAllUsersInCluster(string(clusterRole))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.ScopeCluster)
		delete(query.Filters, iamv1beta1.ResourcesSingularClusterRole)
	} else if globalRole := query.Filters[iamv1beta1.ResourcesSingularGlobalRole]; globalRole != "" {
		users, err = d.listAllUsersByGlobalRole(string(globalRole))
		if err != nil {
			return nil, err
		}
		delete(query.Filters, iamv1beta1.ResourcesSingularGlobalRole)
	} else {
		userList := &iamv1beta1.UserList{}
		if err := d.cache.List(context.Background(), userList,
			client.MatchingLabelsSelector{Selector: query.Selector()}); err != nil {
			return nil, err
		}
		users = make([]*iamv1beta1.User, 0)
		for _, item := range userList.Items {
			users = append(users, item.DeepCopy())
		}
	}

	var result []runtime.Object
	for _, user := range users {
		result = append(result, user)
	}

	return v1alpha3.DefaultList(result, query, d.compare, d.filter), nil
}

func (d *usersGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {

	leftUser, ok := left.(*iamv1beta1.User)
	if !ok {
		return false
	}

	rightUser, ok := right.(*iamv1beta1.User)
	if !ok {
		return false
	}

	return v1alpha3.DefaultObjectMetaCompare(leftUser.ObjectMeta, rightUser.ObjectMeta, field)
}

func (d *usersGetter) filter(object runtime.Object, filter query.Filter) bool {
	user, ok := object.(*iamv1beta1.User)

	if !ok {
		return false
	}

	switch filter.Field {
	case iamv1beta1.FieldEmail:
		return user.Spec.Email == string(filter.Value)
	case iamv1beta1.InGroup:
		return sliceutil.HasString(user.Spec.Groups, string(filter.Value))
	case iamv1beta1.NotInGroup:
		return !sliceutil.HasString(user.Spec.Groups, string(filter.Value))
	default:
		return v1alpha3.DefaultObjectMetaFilter(user.ObjectMeta, filter)
	}
}

func (d *usersGetter) listAllUsersInWorkspace(workspace, role string) ([]*iamv1beta1.User, error) {
	var users []*iamv1beta1.User
	workspaceRoleBindingList := &iamv1beta1.WorkspaceRoleBindingList{}
	if err := d.cache.List(context.Background(), workspaceRoleBindingList,
		client.MatchingLabels{tenantv1beta1.WorkspaceLabel: workspace}); err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range workspaceRoleBindingList.Items {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1beta1.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1beta1.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1beta1.WorkspaceRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInNamespace(namespace, role string) ([]*iamv1beta1.User, error) {
	var users []*iamv1beta1.User

	roleBindingList := &iamv1beta1.RoleBindingList{}
	if err := d.cache.List(context.Background(), roleBindingList, client.InNamespace(namespace)); err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range roleBindingList.Items {
		if role != "" && roleBinding.RoleRef.Name != role {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1beta1.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1beta1.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1beta1.RoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersByGlobalRole(globalRole string) ([]*iamv1beta1.User, error) {
	var users []*iamv1beta1.User
	globalRoleBindingList := &iamv1beta1.GlobalRoleBindingList{}
	if err := d.cache.List(context.Background(), globalRoleBindingList); err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range globalRoleBindingList.Items {
		if roleBinding.RoleRef.Name != globalRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1beta1.ResourceKindUser {

				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1beta1.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1beta1.GlobalRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func (d *usersGetter) listAllUsersInCluster(clusterRole string) ([]*iamv1beta1.User, error) {
	var users []*iamv1beta1.User

	clusterRoleBindingList := &iamv1beta1.WorkspaceRoleBindingList{}
	if err := d.cache.List(context.Background(), clusterRoleBindingList); err != nil {
		klog.Error(err)
		return nil, err
	}

	for _, roleBinding := range clusterRoleBindingList.Items {
		if clusterRole != "" && roleBinding.RoleRef.Name != clusterRole {
			continue
		}
		for _, subject := range roleBinding.Subjects {
			if subject.Kind == iamv1beta1.ResourceKindUser {
				if contains(users, subject.Name) {
					klog.Warningf("conflict role binding found: %s, username:%s", roleBinding.ObjectMeta.String(), subject.Name)
					continue
				}

				obj, err := d.Get("", subject.Name)

				if err != nil {
					if errors.IsNotFound(err) {
						klog.Warningf("orphan subject: %s", subject.String())
						continue
					}
					klog.Error(err)
					return nil, err
				}

				user := obj.(*iamv1beta1.User)
				user = user.DeepCopy()
				if user.Annotations == nil {
					user.Annotations = make(map[string]string, 0)
				}
				user.Annotations[iamv1beta1.ClusterRoleAnnotation] = roleBinding.RoleRef.Name
				users = append(users, user)
			}
		}
	}

	return users, nil
}

func contains(users []*iamv1beta1.User, username string) bool {
	for _, user := range users {
		if user.Name == username {
			return true
		}
	}
	return false
}

package am

import (
	"context"
	"fmt"

	"k8s.io/klog/v2"

	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	tenantv1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	rv1beta1 "kubesphere.io/kubesphere/pkg/models/resources/v1beta1"
)

const (
	emptyNamespace = ""
	groupName      = "iam"
)

type ReadOnlyAccessManagementInterface interface {
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)
	GetGlobalRole(globalRole string) (*iamv1beta1.GlobalRole, error)
	GetClusterRole(name string) (*iamv1beta1.ClusterRole, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	GetNamespaceRole(namespace string, name string) (*iamv1beta1.Role, error)
	ListRoles(namespace string, query *query.Query) (*api.ListResult, error)
	ListWorkspaceRoles(query *query.Query) (*api.ListResult, error)
	GetWorkspaceRole(workspace string, name string) (*iamv1beta1.WorkspaceRole, error)
	ListRoleTemplate(query *query.Query) (*api.ListResult, error)
	GetRoleTemplate(name string) (*iamv1beta1.RoleTemplate, error)
	CreateRoleBinding(namespace string, roleBinding *iamv1beta1.RoleBinding) (*iamv1beta1.RoleBinding, error)
	GetRoleBinding(namespace, name string) (*iamv1beta1.RoleBinding, error)
	ListRoleBindings(username string, groups []string, namespace string) ([]*iamv1beta1.RoleBinding, error)
	ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1beta1.WorkspaceRoleBinding, error)
	ListGlobalRoleBindings(username string) ([]*iamv1beta1.GlobalRoleBinding, error)
	ListClusterRoleBindings(username string) ([]*iamv1beta1.ClusterRoleBinding, error)
	ListCategories(queryParam *query.Query) (*api.ListResult, error)
	GetCategory(name string) (*iamv1beta1.Category, error)
}

type AccessManagementInterface interface {
	ListGlobalRoles(query *query.Query) (*api.ListResult, error)
	CreateOrUpdateGlobalRole(globalRole *iamv1beta1.GlobalRole) (*iamv1beta1.GlobalRole, error)
	PatchGlobalRole(globalRole *iamv1beta1.GlobalRole) (*iamv1beta1.GlobalRole, error)
	DeleteGlobalRole(name string) error
	GetGlobalRole(globalRole string) (*iamv1beta1.GlobalRole, error)

	CreateOrUpdateClusterRole(clusterRole *iamv1beta1.ClusterRole) (*iamv1beta1.ClusterRole, error)
	DeleteClusterRole(name string) error
	GetClusterRole(name string) (*iamv1beta1.ClusterRole, error)
	ListClusterRoles(query *query.Query) (*api.ListResult, error)
	PatchClusterRole(clusterRole *iamv1beta1.ClusterRole) (*iamv1beta1.ClusterRole, error)

	GetNamespaceRole(namespace string, name string) (*iamv1beta1.Role, error)
	CreateOrUpdateNamespaceRole(namespace string, role *iamv1beta1.Role) (*iamv1beta1.Role, error)
	DeleteNamespaceRole(namespace string, name string) error
	PatchNamespaceRole(namespace string, role *iamv1beta1.Role) (*iamv1beta1.Role, error)
	ListRoles(namespace string, query *query.Query) (*api.ListResult, error)

	ListWorkspaceRoles(query *query.Query) (*api.ListResult, error)
	CreateOrUpdateWorkspaceRole(workspace string, workspaceRole *iamv1beta1.WorkspaceRole) (*iamv1beta1.WorkspaceRole, error)
	PatchWorkspaceRole(workspace string, workspaceRole *iamv1beta1.WorkspaceRole) (*iamv1beta1.WorkspaceRole, error)
	DeleteWorkspaceRole(workspace string, name string) error
	GetWorkspaceRole(workspace string, name string) (*iamv1beta1.WorkspaceRole, error)

	ListRoleTemplate(query *query.Query) (*api.ListResult, error)
	GetRoleTemplate(name string) (*iamv1beta1.RoleTemplate, error)
	CreateOrUpdateRoleTemplate(roleTemplate *iamv1beta1.RoleTemplate) (*iamv1beta1.RoleTemplate, error)
	DeleteRoleTemplate(name string) error

	CreateRoleBinding(namespace string, roleBinding *iamv1beta1.RoleBinding) (*iamv1beta1.RoleBinding, error)
	DeleteRoleBinding(namespace, name string) error
	ListRoleBindings(username string, groups []string, namespace string) ([]*iamv1beta1.RoleBinding, error)
	GetRoleBinding(namespace, name string) (*iamv1beta1.RoleBinding, error)

	CreateWorkspaceRoleBinding(workspace string, roleBinding *iamv1beta1.WorkspaceRoleBinding) (*iamv1beta1.WorkspaceRoleBinding, error)
	DeleteWorkspaceRoleBinding(workspaceName, name string) error
	ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1beta1.WorkspaceRoleBinding, error)

	ListGlobalRoleBindings(username string) ([]*iamv1beta1.GlobalRoleBinding, error)
	CreateGlobalRoleBinding(username string, globalRole string) error

	ListClusterRoleBindings(username string) ([]*iamv1beta1.ClusterRoleBinding, error)
	CreateClusterRoleBinding(username string, role string) error

	CreateOrUpdateCategory(category *iamv1beta1.Category) (*iamv1beta1.Category, error)
	DeleteCategory(name string) error
	ListCategories(queryParam *query.Query) (*api.ListResult, error)
	GetCategory(name string) (*iamv1beta1.Category, error)

	//RemoveUserFromWorkspace(username string, workspace string) error
	//CreateNamespaceRoleBinding(username string, namespace string, role string) error
	//RemoveUserFromNamespace(username string, namespace string) error

	//RemoveUserFromCluster(username string) error

	//GetNamespaceControlledWorkspace(namespace string) (string, error)
	//GetDevOpsControlledWorkspace(devops string) (string, error)

	//CreateUserWorkspaceRoleBinding(username string, workspace string, role string) error

	//ListGroupWorkspaceRoleBindings(workspace string, query *query.Query) (*api.ListResult, error)

}

type amOperator struct {
	readOnlyClient rv1beta1.Interface
	client         client.Client
}

func NewReadOnlyOperator(reader rv1beta1.Interface) ReadOnlyAccessManagementInterface {
	return &amOperator{readOnlyClient: reader}
}

func NewOperator(reader rv1beta1.Interface, client client.Client) AccessManagementInterface {
	operator := NewReadOnlyOperator(reader).(*amOperator)
	operator.client = client
	return operator
}

func (am *amOperator) CreateOrUpdateCategory(category *iamv1beta1.Category) (*iamv1beta1.Category, error) {
	err := am.createOrUpdateObj(category)
	if err != nil {
		return nil, err
	}

	return am.GetCategory(category.Name)
}

func (am *amOperator) DeleteCategory(name string) error {
	roleTemplate := &iamv1beta1.Category{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return am.client.Delete(context.Background(), roleTemplate)
}

func (am *amOperator) ListCategories(queryParam *query.Query) (*api.ListResult, error) {
	var (
		result api.ListResult
		list   iamv1beta1.CategoryList
	)

	err := am.readOnlyClient.List(emptyNamespace, queryParam, &list)
	if err != nil {
		return nil, err
	}
	for _, category := range list.Items {
		result.Items = append(result.Items, category)
	}
	result.TotalItems = len(result.Items)

	return &result, nil
}

func (am *amOperator) GetCategory(name string) (*iamv1beta1.Category, error) {
	var roleTemplate iamv1beta1.Category
	err := am.readOnlyClient.Get(emptyNamespace, name, &roleTemplate)
	if err != nil {
		return nil, err
	}
	return &roleTemplate, nil
}

func (am *amOperator) ListRoleTemplate(query *query.Query) (*api.ListResult, error) {
	var (
		result api.ListResult
		list   iamv1beta1.RoleTemplateList
	)

	err := am.readOnlyClient.List(emptyNamespace, query, &list)
	if err != nil {
		return nil, err
	}
	for _, roleTemplate := range list.Items {
		result.Items = append(result.Items, roleTemplate)
	}
	result.TotalItems = len(result.Items)

	return &result, nil
}

func (am *amOperator) GetRoleTemplate(name string) (*iamv1beta1.RoleTemplate, error) {
	var roleTemplate iamv1beta1.RoleTemplate
	err := am.readOnlyClient.Get(emptyNamespace, name, &roleTemplate)
	if err != nil {
		return nil, err
	}
	return &roleTemplate, nil
}

func (am *amOperator) CreateOrUpdateRoleTemplate(roleTemplate *iamv1beta1.RoleTemplate) (*iamv1beta1.RoleTemplate, error) {
	err := am.createOrUpdateObj(roleTemplate)
	if err != nil {
		return nil, err
	}

	return am.GetRoleTemplate(roleTemplate.Name)
}

func (am *amOperator) DeleteRoleTemplate(name string) error {
	roleTemplate := &iamv1beta1.RoleTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return am.client.Delete(context.Background(), roleTemplate)
}

func (am *amOperator) ListGlobalRoles(query *query.Query) (*api.ListResult, error) {
	list := &iamv1beta1.GlobalRoleList{}
	result := &api.ListResult{}
	err := am.readOnlyClient.List(emptyNamespace, query, list)
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		result.Items = append(result.Items, item)
	}
	result.TotalItems = len(result.Items)

	return result, nil
}

func (am *amOperator) CreateOrUpdateGlobalRole(globalRole *iamv1beta1.GlobalRole) (*iamv1beta1.GlobalRole, error) {
	roleTemplateNames, rules, err := am.aggregateRules(globalRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}

	globalRole.Rules = rules
	globalRole.AggregationRoleTemplates.TemplateNames = roleTemplateNames

	err = am.createOrUpdateObj(globalRole)
	if err != nil {
		return nil, err
	}

	return am.GetGlobalRole(globalRole.Name)
}

func (am *amOperator) PatchGlobalRole(globalRole *iamv1beta1.GlobalRole) (*iamv1beta1.GlobalRole, error) {
	if globalRole.AggregationRoleTemplates.Selector.Size() != 0 {
		return nil, errors.NewBadRequest("cannot use labels to patch role")
	}

	oldRole, err := am.GetClusterRole(globalRole.Name)
	if err != nil {
		return nil, err
	}

	templateNames, rules, err := am.aggregateRules(globalRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}
	globalRole.Rules = rules
	globalRole.AggregationRoleTemplates.TemplateNames = templateNames

	err = am.client.Patch(context.Background(), globalRole, client.MergeFrom(oldRole))
	if err != nil {
		return nil, err
	}

	return am.GetGlobalRole(globalRole.Name)
}

func (am *amOperator) DeleteGlobalRole(name string) error {
	globalRole := &iamv1beta1.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return am.client.Delete(context.Background(), globalRole)
}

func (am *amOperator) GetGlobalRole(name string) (*iamv1beta1.GlobalRole, error) {
	var role iamv1beta1.GlobalRole
	err := am.readOnlyClient.Get(emptyNamespace, name, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (am *amOperator) CreateOrUpdateClusterRole(clusterRole *iamv1beta1.ClusterRole) (*iamv1beta1.ClusterRole, error) {
	roleTemplateNames, rules, err := am.aggregateRules(clusterRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}

	clusterRole.Rules = rules
	clusterRole.AggregationRoleTemplates.TemplateNames = roleTemplateNames

	err = am.createOrUpdateObj(clusterRole)
	if err != nil {
		return nil, err
	}

	return am.GetClusterRole(clusterRole.Name)
}

func (am *amOperator) DeleteClusterRole(name string) error {
	clusterRole := &iamv1beta1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return am.client.Delete(context.Background(), clusterRole)
}

func (am *amOperator) GetClusterRole(name string) (*iamv1beta1.ClusterRole, error) {
	var clusterRole iamv1beta1.ClusterRole
	err := am.readOnlyClient.Get(emptyNamespace, name, &clusterRole)
	if err != nil {
		return nil, err
	}
	return &clusterRole, nil
}

func (am *amOperator) ListClusterRoles(query *query.Query) (*api.ListResult, error) {
	list := &iamv1beta1.ClusterRoleList{}
	result := &api.ListResult{}
	err := am.readOnlyClient.List(emptyNamespace, query, list)
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		result.Items = append(result.Items, item)
	}
	result.TotalItems = len(result.Items)

	return result, nil
}

func (am *amOperator) PatchClusterRole(clusterRole *iamv1beta1.ClusterRole) (*iamv1beta1.ClusterRole, error) {
	if clusterRole.AggregationRoleTemplates.Selector.Size() != 0 {
		return nil, errors.NewBadRequest("cannot use labels to patch role")
	}

	oldRole, err := am.GetClusterRole(clusterRole.Name)
	if err != nil {
		return nil, err
	}

	templateNames, rules, err := am.aggregateRules(clusterRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}
	clusterRole.Rules = rules
	clusterRole.AggregationRoleTemplates.TemplateNames = templateNames

	err = am.client.Patch(context.Background(), clusterRole, client.MergeFrom(oldRole))
	if err != nil {
		return nil, err
	}

	return am.GetClusterRole(clusterRole.Name)
}

func (am *amOperator) GetNamespaceRole(namespace string, name string) (*iamv1beta1.Role, error) {
	var role iamv1beta1.Role
	err := am.readOnlyClient.Get(namespace, name, &role)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (am *amOperator) CreateOrUpdateNamespaceRole(namespace string, role *iamv1beta1.Role) (*iamv1beta1.Role, error) {
	roleTemplateNames, rules, err := am.aggregateRules(role.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}

	role.Rules = rules
	role.AggregationRoleTemplates.TemplateNames = roleTemplateNames

	err = am.createOrUpdateObj(role)
	if err != nil {
		return nil, err
	}

	return am.GetNamespaceRole(namespace, role.Name)
}

func (am *amOperator) DeleteNamespaceRole(namespace string, name string) error {
	role := &iamv1beta1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return am.client.Delete(context.Background(), role)

}

func (am *amOperator) PatchNamespaceRole(namespace string, role *iamv1beta1.Role) (*iamv1beta1.Role, error) {
	if role.AggregationRoleTemplates.Selector.Size() != 0 {
		return nil, errors.NewBadRequest("cannot use labels to patch role")
	}

	oldRole, err := am.GetNamespaceRole(namespace, role.Name)
	if err != nil {
		return nil, err
	}

	templateNames, rules, err := am.aggregateRules(role.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}
	role.Rules = rules
	role.AggregationRoleTemplates.TemplateNames = templateNames

	err = am.client.Patch(context.Background(), role, client.MergeFrom(oldRole))
	if err != nil {
		return nil, err
	}

	return am.GetNamespaceRole(namespace, role.Name)
}

func (am *amOperator) ListRoles(namespace string, query *query.Query) (*api.ListResult, error) {
	list := &iamv1beta1.RoleList{}
	result := &api.ListResult{}
	err := am.readOnlyClient.List(namespace, query, list)
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		result.Items = append(result.Items, item)
	}
	result.TotalItems = len(result.Items)

	return result, nil
}

func (am *amOperator) ListWorkspaceRoles(query *query.Query) (*api.ListResult, error) {
	list := &iamv1beta1.RoleList{}
	result := &api.ListResult{}
	err := am.readOnlyClient.List(emptyNamespace, query, list)
	if err != nil {
		return nil, err
	}

	for _, item := range list.Items {
		result.Items = append(result.Items, item)
	}
	result.TotalItems = len(result.Items)

	return result, nil
}

func (am *amOperator) CreateOrUpdateWorkspaceRole(workspace string, workspaceRole *iamv1beta1.WorkspaceRole) (*iamv1beta1.WorkspaceRole, error) {
	if workspaceRole.Labels == nil {
		workspaceRole.Labels = make(map[string]string, 0)
	}
	workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] = workspace

	roleTemplateNames, rules, err := am.aggregateRules(workspaceRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}

	workspaceRole.Rules = rules
	workspaceRole.AggregationRoleTemplates.TemplateNames = roleTemplateNames

	err = am.createOrUpdateObj(workspaceRole)
	if err != nil {
		return nil, err
	}

	return am.GetWorkspaceRole(emptyNamespace, workspaceRole.Name)
}

func (am *amOperator) PatchWorkspaceRole(workspace string, workspaceRole *iamv1beta1.WorkspaceRole) (*iamv1beta1.WorkspaceRole, error) {
	if workspaceRole.AggregationRoleTemplates.Selector.Size() != 0 {
		return nil, errors.NewBadRequest("cannot use labels to patch role")
	}

	oldRole, err := am.GetWorkspaceRole(workspace, workspaceRole.Name)
	if err != nil {
		return nil, err
	}

	// workspace label cannot be override
	if workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] != "" {
		workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	}

	templateNames, rules, err := am.aggregateRules(workspaceRole.AggregationRoleTemplates)
	if err != nil {
		return nil, err
	}
	workspaceRole.Rules = rules
	workspaceRole.AggregationRoleTemplates.TemplateNames = templateNames

	err = am.client.Patch(context.Background(), workspaceRole, client.MergeFrom(oldRole))
	if err != nil {
		return nil, err
	}

	return am.GetWorkspaceRole(workspace, workspaceRole.Name)
}

func (am *amOperator) DeleteWorkspaceRole(_ string, name string) error {
	globalRole := &iamv1beta1.WorkspaceRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	return am.client.Delete(context.Background(), globalRole)
}

func (am *amOperator) GetWorkspaceRole(workspace string, name string) (*iamv1beta1.WorkspaceRole, error) {
	var workspaceRole iamv1beta1.WorkspaceRole
	err := am.readOnlyClient.Get(emptyNamespace, name, &workspaceRole)
	if err != nil {
		return nil, err
	}

	if workspace != "" && workspaceRole.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(iamv1beta1.Resource("workspacerole"), name)
		klog.Error(err)
		return nil, err
	}
	return &workspaceRole, nil
}

func (am *amOperator) CreateRoleBinding(namespace string, roleBinding *iamv1beta1.RoleBinding) (*iamv1beta1.RoleBinding, error) {
	_, err := am.GetNamespaceRole(namespace, roleBinding.RoleRef.Name)
	if err != nil {
		return nil, err
	}

	if len(roleBinding.Subjects) == 0 {
		err := errors.NewNotFound(iamv1beta1.Resource("rolebinding"), roleBinding.Name)
		return nil, err
	}

	roleBinding.GenerateName = fmt.Sprintf("%s-%s-", roleBinding.Subjects[0].Name, roleBinding.RoleRef.Name)

	if roleBinding.Labels == nil {
		roleBinding.Labels = map[string]string{}
	}

	if roleBinding.Subjects[0].Kind == rbacv1.GroupKind {
		roleBinding.Labels[iamv1beta1.GroupReferenceLabel] = roleBinding.Subjects[0].Name
	} else if roleBinding.Subjects[0].Kind == rbacv1.UserKind {
		roleBinding.Labels[iamv1beta1.UserReferenceLabel] = roleBinding.Subjects[0].Name
	}

	err = am.client.Create(context.Background(), roleBinding)
	if err != nil {
		return nil, err
	}

	return am.GetRoleBinding(namespace, roleBinding.Name)
}

func (am *amOperator) GetRoleBinding(namespace, name string) (*iamv1beta1.RoleBinding, error) {
	rolebinding := &iamv1beta1.RoleBinding{}

	err := am.readOnlyClient.Get(namespace, name, rolebinding)
	if err != nil {
		return nil, err
	}

	return rolebinding, nil
}

func (am *amOperator) DeleteRoleBinding(namespace, name string) error {
	rolebinding := &iamv1beta1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return am.client.Delete(context.Background(), rolebinding)
}

func (am *amOperator) ListRoleBindings(username string, groups []string, namespace string) ([]*iamv1beta1.RoleBinding, error) {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) CreateWorkspaceRoleBinding(workspace string, roleBinding *iamv1beta1.WorkspaceRoleBinding) (*iamv1beta1.WorkspaceRoleBinding, error) {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) DeleteWorkspaceRoleBinding(workspaceName, name string) error {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) ListWorkspaceRoleBindings(username string, groups []string, workspace string) ([]*iamv1beta1.WorkspaceRoleBinding, error) {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) ListGlobalRoleBindings(username string) ([]*iamv1beta1.GlobalRoleBinding, error) {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) CreateGlobalRoleBinding(username string, globalRole string) error {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) ListClusterRoleBindings(username string) ([]*iamv1beta1.ClusterRoleBinding, error) {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) CreateClusterRoleBinding(username string, role string) error {
	//TODO implement me
	panic("implement me")
}

func (am *amOperator) aggregateRules(aggregation iamv1beta1.AggregationRoleTemplates) ([]string, []rbacv1.PolicyRule, error) {
	templateNames := aggregation.TemplateNames
	rules := make([]rbacv1.PolicyRule, 0)

	if aggregation.Selector.Size() != 0 {
		list := &iamv1beta1.RoleTemplateList{}
		err := am.readOnlyClient.List(emptyNamespace, &query.Query{LabelSelector: aggregation.Selector.String()}, list)
		if err != nil {
			return nil, nil, err
		}
		newTemplateName := make([]string, 0)
		for _, roleTemplate := range list.Items {
			rules = append(rules, roleTemplate.Spec.Rules...)
			newTemplateName = append(newTemplateName, roleTemplate.Name)
		}
		// override
		templateNames = newTemplateName
	} else {
		for _, name := range templateNames {
			roleTemplate, err := am.GetRoleTemplate(name)
			if err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return nil, nil, err
			}
			rules = append(rules, roleTemplate.Spec.Rules...)
		}
	}

	return templateNames, rules, nil
}

func (am *amOperator) createOrUpdateObj(obj client.Object) error {
	var exist bool
	if err := am.readOnlyClient.Get(obj.GetNamespace(), obj.GetName(), obj); err != nil {
		if errors.IsNotFound(err) {
			exist = false
		} else {
			return err
		}
	}

	var err error
	if exist {
		err = am.client.Update(context.Background(), obj)
	} else {
		err = am.client.Create(context.Background(), obj)
	}
	return err
}

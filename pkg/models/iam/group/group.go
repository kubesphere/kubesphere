/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package group

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Masterminds/semver/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	iamv1beta1 "kubesphere.io/api/iam/v1beta1"
	"kubesphere.io/api/tenant/v1beta1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
)

type GroupOperator interface {
	ListGroups(workspace string, queryParam *query.Query) (*api.ListResult, error)
	CreateGroup(workspace string, namespace *iamv1beta1.Group) (*iamv1beta1.Group, error)
	DescribeGroup(workspace, group string) (*iamv1beta1.Group, error)
	DeleteGroup(workspace, group string) error
	UpdateGroup(workspace string, group *iamv1beta1.Group) (*iamv1beta1.Group, error)
	PatchGroup(workspace string, group *iamv1beta1.Group) (*iamv1beta1.Group, error)
	DeleteGroupBinding(workspace, name string) error
	CreateGroupBinding(workspace, groupName, userName string) (*iamv1beta1.GroupBinding, error)
	ListGroupBindings(workspace string, queryParam *query.Query) (*api.ListResult, error)
}

type groupOperator struct {
	client         runtimeclient.Client
	resourceGetter *resourcesv1alpha3.Getter
}

func New(cacheClient runtimeclient.Client, k8sVersion *semver.Version) GroupOperator {
	return &groupOperator{
		resourceGetter: resourcesv1alpha3.NewResourceGetter(cacheClient, k8sVersion),
		client:         cacheClient,
	}
}

func (t *groupOperator) ListGroups(workspace string, queryParam *query.Query) (*api.ListResult, error) {

	if workspace != "" {
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", v1beta1.WorkspaceLabel, workspace))
	}

	result, err := t.resourceGetter.List("groups", "", queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// CreateGroup adds a workspace label to group which indicates group is under the workspace
func (t *groupOperator) CreateGroup(workspace string, group *iamv1beta1.Group) (*iamv1beta1.Group, error) {

	if group.GenerateName == "" {
		err := errors.NewInvalid(iamv1beta1.SchemeGroupVersion.WithKind(iamv1beta1.ResourcePluralGroup).GroupKind(),
			"", []*field.Error{field.Required(field.NewPath("metadata.generateName"), "generateName is required")})
		klog.Error(err)
		return nil, err
	}
	// generateName is used as displayName
	// ensure generateName is unique in workspace scope
	if unique, err := t.isGenerateNameUnique(workspace, group.GenerateName); err != nil {
		return nil, err
	} else if !unique {
		err = errors.NewConflict(iamv1beta1.Resource(iamv1beta1.ResourcePluralGroup),
			group.GenerateName, fmt.Errorf("a group named %s already exists in the workspace", group.GenerateName))
		klog.Error(err)
		return nil, err
	}

	group = labelGroupWithWorkspaceName(group, workspace)
	return group, t.client.Create(context.Background(), group)
}

func (t *groupOperator) isGenerateNameUnique(workspace, generateName string) (bool, error) {

	result, err := t.ListGroups(workspace, query.New())

	if err != nil {
		klog.Error(err)
		return false, err
	}
	for _, obj := range result.Items {
		g := obj.(*iamv1beta1.Group)
		if g.GenerateName == generateName {
			return false, err
		}
	}
	return true, nil
}

func (t *groupOperator) DescribeGroup(workspace, group string) (*iamv1beta1.Group, error) {
	obj, err := t.resourceGetter.Get("groups", "", group)
	if err != nil {
		return nil, err
	}
	ns := obj.(*iamv1beta1.Group)
	if ns.Labels[v1beta1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(corev1.Resource("group"), group)
		klog.Error(err)
		return nil, err
	}
	return ns, nil
}

func (t *groupOperator) DeleteGroup(workspace, groupName string) error {
	group, err := t.DescribeGroup(workspace, groupName)
	if err != nil {
		return err
	}
	return t.client.Delete(context.Background(), group, &runtimeclient.DeleteOptions{GracePeriodSeconds: ptr.To[int64](0)})
}

func (t *groupOperator) UpdateGroup(workspace string, group *iamv1beta1.Group) (*iamv1beta1.Group, error) {
	_, err := t.DescribeGroup(workspace, group.Name)
	if err != nil {
		return nil, err
	}
	group = labelGroupWithWorkspaceName(group, workspace)
	return group, t.client.Update(context.Background(), group)
}

func (t *groupOperator) PatchGroup(workspace string, group *iamv1beta1.Group) (*iamv1beta1.Group, error) {
	group, err := t.DescribeGroup(workspace, group.Name)
	if err != nil {
		return nil, err
	}
	if group.Labels != nil {
		group.Labels[v1beta1.WorkspaceLabel] = workspace
	}
	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	return group, t.client.Patch(context.Background(), group, runtimeclient.RawPatch(types.MergePatchType, data))
}

func (t *groupOperator) DeleteGroupBinding(workspace, name string) error {
	groupBinding := &iamv1beta1.GroupBinding{}
	if err := t.client.Get(context.Background(), types.NamespacedName{Name: name}, groupBinding); err != nil {
		return err
	}
	if groupBinding.Labels[v1beta1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(corev1.Resource("groupbindings"), name)
		klog.Error(err)
		return err
	}

	return t.client.Delete(context.Background(), groupBinding, &runtimeclient.DeleteOptions{GracePeriodSeconds: ptr.To[int64](0)})
}

func (t *groupOperator) CreateGroupBinding(workspace, groupName, userName string) (*iamv1beta1.GroupBinding, error) {

	groupBinding := &iamv1beta1.GroupBinding{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-%s-", groupName, userName),
			Labels: map[string]string{
				iamv1beta1.UserReferenceLabel:  userName,
				iamv1beta1.GroupReferenceLabel: groupName,
				v1beta1.WorkspaceLabel:         workspace,
			},
		},
		Users: []string{userName},
		GroupRef: iamv1beta1.GroupRef{
			APIGroup: iamv1beta1.SchemeGroupVersion.Group,
			Kind:     iamv1beta1.ResourcePluralGroup,
			Name:     groupName,
		},
	}

	return groupBinding, t.client.Create(context.Background(), groupBinding)
}

func (t *groupOperator) ListGroupBindings(workspace string, query *query.Query) (*api.ListResult, error) {

	lableSelector, err := labels.ConvertSelectorToLabelsMap(query.LabelSelector)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	// workspace resources must be filtered by workspace
	wsSelector := labels.Set{v1beta1.WorkspaceLabel: workspace}
	query.LabelSelector = labels.Merge(lableSelector, wsSelector).String()

	result, err := t.resourceGetter.List("groupbindings", "", query)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// labelGroupWithWorkspaceName adds a kubesphere.io/workspace=[workspaceName] label to namespace which
// indicates namespace is under the workspace
func labelGroupWithWorkspaceName(namespace *iamv1beta1.Group, workspaceName string) *iamv1beta1.Group {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[v1beta1.WorkspaceLabel] = workspaceName // label namespace with workspace name

	return namespace
}

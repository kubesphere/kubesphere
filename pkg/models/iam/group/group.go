/*
Copyright 2020 KubeSphere Authors

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

package group

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/api"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	kubesphere "kubesphere.io/kubesphere/pkg/client/clientset/versioned"
	"kubesphere.io/kubesphere/pkg/informers"
	resourcesv1alpha3 "kubesphere.io/kubesphere/pkg/models/resources/v1alpha3/resource"
)

type GroupOperator interface {
	ListGroups(workspace string, queryParam *query.Query) (*api.ListResult, error)
	CreateGroup(workspace string, namespace *iamv1alpha2.Group) (*iamv1alpha2.Group, error)
	DescribeGroup(workspace, group string) (*iamv1alpha2.Group, error)
	DeleteGroup(workspace, group string) error
	UpdateGroup(workspace string, group *iamv1alpha2.Group) (*iamv1alpha2.Group, error)
	PatchGroup(workspace string, group *iamv1alpha2.Group) (*iamv1alpha2.Group, error)
	DeleteGroupBinding(workspace, name string) error
	CreateGroupBinding(workspace, groupName, userName string) error
	ListGroupBindings(workspace, group string, queryParam *query.Query) (*api.ListResult, error)
}

type groupOperator struct {
	k8sclient      kubernetes.Interface
	ksclient       kubesphere.Interface
	resourceGetter *resourcesv1alpha3.ResourceGetter
}

func New(informers informers.InformerFactory, ksclient kubesphere.Interface, k8sclient kubernetes.Interface) GroupOperator {
	return &groupOperator{
		resourceGetter: resourcesv1alpha3.NewResourceGetter(informers),
		k8sclient:      k8sclient,
		ksclient:       ksclient,
	}
}

func (t *groupOperator) ListGroups(workspace string, queryParam *query.Query) (*api.ListResult, error) {

	if workspace != "" {
		// filter by workspace
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", tenantv1alpha1.WorkspaceLabel, workspace))
	}

	result, err := t.resourceGetter.List("groups", "", queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// CreateGroup adds a workspace label to group which indicates group is under the workspace
func (t *groupOperator) CreateGroup(workspace string, namespace *iamv1alpha2.Group) (*iamv1alpha2.Group, error) {
	return t.ksclient.IamV1alpha2().Groups().Create(labelGroupWithWorkspaceName(namespace, workspace))
}

func (t *groupOperator) DescribeGroup(workspace, group string) (*iamv1alpha2.Group, error) {
	obj, err := t.resourceGetter.Get("groups", "", group)
	if err != nil {
		return nil, err
	}
	ns := obj.(*iamv1alpha2.Group)
	if ns.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(corev1.Resource("group"), group)
		klog.Error(err)
		return nil, err
	}
	return ns, nil
}

func (t *groupOperator) DeleteGroup(workspace, group string) error {
	_, err := t.DescribeGroup(workspace, group)
	if err != nil {
		return err
	}
	return t.ksclient.IamV1alpha2().Groups().Delete(group, metav1.NewDeleteOptions(0))
}

func (t *groupOperator) UpdateGroup(workspace string, group *iamv1alpha2.Group) (*iamv1alpha2.Group, error) {
	_, err := t.DescribeGroup(workspace, group.Name)
	if err != nil {
		return nil, err
	}
	group = labelGroupWithWorkspaceName(group, workspace)
	return t.ksclient.IamV1alpha2().Groups().Update(group)
}

func (t *groupOperator) PatchGroup(workspace string, group *iamv1alpha2.Group) (*iamv1alpha2.Group, error) {
	_, err := t.DescribeGroup(workspace, group.Name)
	if err != nil {
		return nil, err
	}
	if group.Labels != nil {
		group.Labels[tenantv1alpha1.WorkspaceLabel] = workspace
	}
	data, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	return t.ksclient.IamV1alpha2().Groups().Patch(group.Name, types.MergePatchType, data)
}

func (t *groupOperator) DeleteGroupBinding(workspace, name string) error {
	obj, err := t.resourceGetter.Get("groupbindings", "", name)
	if err != nil {
		return err
	}
	ns := obj.(*iamv1alpha2.GroupBinding)
	if ns.Labels[tenantv1alpha1.WorkspaceLabel] != workspace {
		err := errors.NewNotFound(corev1.Resource("groupbinding"), name)
		klog.Error(err)
		return err
	}

	return t.ksclient.IamV1alpha2().GroupBindings().Delete(name, metav1.NewDeleteOptions(0))
}

func (t *groupOperator) CreateGroupBinding(workspace, groupName, userName string) error {

	groupBinding := iamv1alpha2.GroupBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", groupName, userName),
			Labels: map[string]string{
				iamv1alpha2.UserReferenceLabel:  userName,
				iamv1alpha2.GroupReferenceLabel: groupName,
				tenantv1alpha1.WorkspaceLabel:   workspace,
			},
		},
		Users: []string{userName},
		GroupRef: iamv1alpha2.GroupRef{
			APIGroup: iamv1alpha2.SchemeGroupVersion.Group,
			Kind:     iamv1alpha2.ResourcePluralGroup,
			Name:     groupName,
		},
	}

	if _, err := t.ksclient.IamV1alpha2().GroupBindings().Create(&groupBinding); err != nil {
		return err
	}

	return nil
}

func (t *groupOperator) ListGroupBindings(workspace, group string, queryParam *query.Query) (*api.ListResult, error) {

	if group != "" && workspace != "" {
		// filter by group
		queryParam.Filters[query.FieldLabel] = query.Value(fmt.Sprintf("%s=%s", iamv1alpha2.GroupReferenceLabel, group))
	}

	result, err := t.resourceGetter.List("groupbindings", "", queryParam)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	return result, nil
}

// labelGroupWithWorkspaceName adds a kubesphere.io/workspace=[workspaceName] label to namespace which
// indicates namespace is under the workspace
func labelGroupWithWorkspaceName(namespace *iamv1alpha2.Group, workspaceName string) *iamv1alpha2.Group {
	if namespace.Labels == nil {
		namespace.Labels = make(map[string]string, 0)
	}

	namespace.Labels[tenantv1alpha1.WorkspaceLabel] = workspaceName // label namespace with workspace name

	return namespace
}

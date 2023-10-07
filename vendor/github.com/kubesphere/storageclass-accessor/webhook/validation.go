package webhook

import (
	"context"
	"fmt"

	"github.com/kubesphere/storageclass-accessor/client/apis/accessor/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	workspacev1alpha1 "kubesphere.io/api/tenant/v1alpha1"
)

func (a *Admitter) validateNameSpace(ctx context.Context, reqResource ReqInfo, accessor *v1alpha1.Accessor) error {
	klog.Infof("start validating namespace: %s", reqResource.namespace)
	ns, err := a.getNameSpace(ctx, reqResource.namespace)
	if err != nil {
		klog.ErrorS(err, "get namespace failed", "namespace", reqResource.namespace)
		return err
	}
	fieldPass := nsMatchField(ns, accessor.Spec.NameSpaceSelector.FieldSelector)
	labelPass := matchLabel(ns.Labels, accessor.Spec.NameSpaceSelector.LabelSelector)
	if fieldPass && labelPass {
		return nil
	}

	err = fmt.Errorf("%s %s is not allowed %s in the namespace: %s", reqResource.resource, reqResource.name, reqResource.operator, reqResource.namespace)
	klog.ErrorS(err, "validate namespace failed")
	return err
}

func (a *Admitter) validateWorkSpace(ctx context.Context, reqResource ReqInfo, accessor *v1alpha1.Accessor) error {
	klog.Infof("start validating workspace for namespace: %s", reqResource.namespace)

	ns, err := a.getNameSpace(ctx, reqResource.namespace)
	if err != nil {
		klog.ErrorS(err, "get namespace failed", "namespace", reqResource.namespace)
		return err
	}
	if wsName, ok := ns.Labels["kubesphere.io/workspace"]; ok {
		klog.Infof("namespace %s is in workspace %s", reqResource.namespace, wsName)
		var ws *workspacev1alpha1.Workspace
		ws, err = a.getWorkSpace(ctx, wsName)
		if err != nil {
			klog.ErrorS(err, "failed to get the workspace", "workspace", wsName)
			return err
		}
		fieldPass := wsMatchField(ws, accessor.Spec.WorkSpaceSelector.FieldSelector)
		labelPass := matchLabel(ws.Labels, accessor.Spec.WorkSpaceSelector.LabelSelector)
		if fieldPass && labelPass {
			return nil
		}

		err = fmt.Errorf("%s %s is not allowed %s in the workspace: %s", reqResource.resource, reqResource.name, reqResource.operator, wsName)
		klog.ErrorS(err, "validate workspace failed", "workspace", wsName)
		return err
	}
	klog.Infof("namespace %s has no workspace information, skipped", reqResource.namespace)
	return nil
}

func (a *Admitter) getNameSpace(ctx context.Context, nameSpaceName string) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{}
	err := a.client.Get(ctx, types.NamespacedName{Name: nameSpaceName}, ns)
	return ns, err
}

func (a *Admitter) getAccessors(ctx context.Context, storageClassName string) ([]v1alpha1.Accessor, error) {
	accessorList := &v1alpha1.AccessorList{}
	err := a.client.List(ctx, accessorList)
	if err != nil {
		klog.ErrorS(err, "failed to list accessors for storage class", "storageClassName", storageClassName)
		// TODO If not found , pass or not?
		return nil, err
	}
	list := make([]v1alpha1.Accessor, 0)
	for _, accessor := range accessorList.Items {
		if accessor.Spec.StorageClassName == storageClassName {
			list = append(list, accessor)
		}
	}
	return list, nil
}

func (a *Admitter) getWorkSpace(ctx context.Context, workspaceName string) (*workspacev1alpha1.Workspace, error) {
	workspace := &workspacev1alpha1.Workspace{}
	err := a.client.Get(ctx, types.NamespacedName{Name: workspaceName}, workspace)
	return workspace, err
}

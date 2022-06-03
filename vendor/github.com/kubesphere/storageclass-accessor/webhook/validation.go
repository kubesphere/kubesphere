package webhook

import (
	"context"
	"fmt"
	"github.com/kubesphere/storageclass-accessor/client/apis/accessor/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	workspacev1alpha1 "kubesphere.io/api/tenant/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func validateNameSpace(reqResource ReqInfo, accessor *v1alpha1.Accessor) error {
	klog.Info("start validate namespace")
	//accessor, err := getAccessor()
	ns, err := getNameSpace(reqResource.namespace)
	if err != nil {
		klog.Error(err)
		return err
	}
	var fieldPass, labelPass bool
	fieldPass = matchField(ns, accessor.Spec.NameSpaceSelector.FieldSelector)
	labelPass = matchLabel(ns.Labels, accessor.Spec.NameSpaceSelector.LabelSelector)
	if fieldPass && labelPass {
		return nil
	}

	klog.Error(fmt.Sprintf("%s %s does not allowed %s in the namespace: %s", reqResource.resource, reqResource.name, reqResource.operator, reqResource.namespace))
	return fmt.Errorf("The storageClass: %s does not allowed %s %s %s in the namespace: %s ", reqResource.storageClassName, reqResource.operator, reqResource.resource, reqResource.name, reqResource.namespace)
}

func validateWorkSpace(reqResource ReqInfo, accessor *v1alpha1.Accessor) error {
	klog.Info("start validate workspace")

	ns, err := getNameSpace(reqResource.namespace)
	if err != nil {
		klog.Error(err)
		return err
	}
	if wsName, ok := ns.Labels["kubesphere.io/workspace"]; ok {
		var ws *workspacev1alpha1.Workspace
		ws, err = getWorkSpace(wsName)
		if err != nil {
			klog.Error("Cannot get the workspace")
		}
		var fieldPass, labelPass bool
		fieldPass = wsMatchField(ws, accessor.Spec.WorkSpaceSelector.FieldSelector)
		labelPass = matchLabel(ns.Labels, accessor.Spec.WorkSpaceSelector.LabelSelector)
		if fieldPass && labelPass {
			return nil
		}

		klog.Error(fmt.Sprintf("%s %s does not allowed %s in the workspace: %s", reqResource.resource, reqResource.name, reqResource.operator, wsName))
		return fmt.Errorf("The storageClass: %s does not allowed %s %s %s in the workspace: %s ", reqResource.storageClassName, reqResource.operator, reqResource.resource, reqResource.name, wsName)
	}
	klog.Info("Unable to get workspace information, skipped.")
	return nil
}

func getNameSpace(nameSpaceName string) (*corev1.Namespace, error) {
	nsClient, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		return nil, err
	}
	ns := &corev1.Namespace{}
	err = nsClient.Get(context.Background(), types.NamespacedName{Namespace: "", Name: nameSpaceName}, ns)
	if err != nil {
		klog.Error("client get namespace failed, err:", err)
		return nil, err
	}
	return ns, nil
}

func getAccessors(storageClassName string) ([]*v1alpha1.Accessor, error) {
	// get config
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	var cli client.Client
	opts := client.Options{}
	scheme := runtime.NewScheme()
	_ = v1alpha1.AddToScheme(scheme)
	opts.Scheme = scheme
	cli, err = client.New(cfg, opts)
	if err != nil {
		return nil, err
	}
	accessorList := &v1alpha1.AccessorList{}

	var listOpt []client.ListOption
	err = cli.List(context.Background(), accessorList, listOpt...)
	if err != nil {
		// TODO If not found , pass or not?
		return nil, err
	}
	list := make([]*v1alpha1.Accessor, 0)
	for _, accessor := range accessorList.Items {
		if accessor.Spec.StorageClassName == storageClassName {
			list = append(list, &accessor)
		}
	}
	return list, nil
}

func getWorkSpace(workspaceName string) (*workspacev1alpha1.Workspace, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	var cli client.Client
	opts := client.Options{}
	scheme := runtime.NewScheme()
	_ = workspacev1alpha1.AddToScheme(scheme)
	opts.Scheme = scheme
	cli, err = client.New(cfg, opts)
	if err != nil {
		return nil, err
	}
	workspace := &workspacev1alpha1.Workspace{}
	err = cli.Get(context.Background(), types.NamespacedName{Namespace: "", Name: workspaceName}, workspace)
	if err != nil {
		klog.Error("can't get the workspace by name, err:", err)
	}
	return workspace, err
}

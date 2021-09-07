package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"kubesphere.io/api/manifest/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *ManifestReconciler) patchCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(resource.Spec.CustomResource), nil, obj)
	if err != nil {
		klog.Errorf("get gvk error: %s", err.Error())
		return err
	}

	obj.SetName(resource.Name)
	obj.SetNamespace(resource.Namespace)

	err = r.Client.Patch(ctx, obj, client.Merge)
	if err != nil {
		klog.Info(err.Error())
		return err
	}
	return nil
}

func (r *ManifestReconciler) deleteCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(resource.Spec.CustomResource), nil, obj)
	if err != nil {
		klog.Errorf("get gvk error: %s", err.Error())
		return err
	}
	err = r.Delete(ctx, obj)
	return err
}

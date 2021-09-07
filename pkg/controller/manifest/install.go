package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"kubesphere.io/api/manifest/v1alpha1"
	"time"
)

var decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

func (r *ManifestReconciler) installCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	obj := &unstructured.Unstructured{}
	_, _, err := decUnstructured.Decode([]byte(resource.Spec.CustomResource), nil, obj)
	if err != nil {
		klog.Errorf("get gvk error: %s", err.Error())
		return err
	}
	obj.SetName(resource.Name)
	obj.SetNamespace(resource.Namespace)
	err = r.Create(ctx, obj)
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)
	var clusterStatus string
	err = r.Get(ctx, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)

	if err != nil {
		klog.Error(err.Error())
		return err
	}
	statusMap, ok := obj.Object["status"].(map[string]interface{})
	if ok {
		clusterStatus = statusMap["status"].(string)
	} else {
		clusterStatus = v1alpha1.ClusterStatusUnknown
	}
	resource.Status.Status = clusterStatus
	switch resource.Kind {
	case v1alpha1.DBTypeClickHouse:
		resource.Spec.Application = v1alpha1.ClusterAppTypeClickHouse
	case v1alpha1.DBTypePostgreSQL:
		resource.Spec.Application = v1alpha1.ClusterAPPTypePostgreSQL
	case v1alpha1.DBTypeMysql:
		resource.Spec.Application = v1alpha1.ClusterAPPTypeMySQL
	default:
		resource.Spec.Application = ""
	}
	err = r.Client.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.Status = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}
	return nil
}

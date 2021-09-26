/*
Copyright 2021.

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

package controllers

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const CheckTime = 1 * time.Minute

var (
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	finalizer       = "finalizers.radondb.com/customresource"
)

// ManifestReconciler reconciles a Manifest object
type ManifestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=application.kubesphere.io,resources=manifests,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=application.kubesphere.io,resources=manifests/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=application.kubesphere.io,resources=manifests/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Manifest object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *ManifestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var customResource  v1alpha1.Manifest
	if err := r.Get(ctx, req.NamespacedName, &customResource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if customResource.Status.Status == "" {
		customResource.Status.Status = v1alpha1.Creating
		customResource.Status.LastUpdate = metav1.Now()
		err := r.Status().Update(ctx, &customResource)
		return reconcile.Result{}, err
	}

	if customResource.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted
		if !sliceutil.HasString(customResource.ObjectMeta.Finalizers, finalizer) {
			customResource.ObjectMeta.Finalizers = append(customResource.ObjectMeta.Finalizers, finalizer)
			err := r.Update(ctx, &customResource)
			return reconcile.Result{}, err
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(customResource.ObjectMeta.Finalizers, finalizer) {
			err := r.deleteCluster(ctx, &customResource)
			if err != nil {
				klog.Errorf("delete database cluster error: %s", client.IgnoreNotFound(err).Error())
			}
			customResource.ObjectMeta.Finalizers = sliceutil.RemoveString(customResource.ObjectMeta.Finalizers, func(item string) bool {
				if item == finalizer {
					return true
				}
				return false
			})
			err = r.Update(ctx, &customResource)
			return reconcile.Result{}, err
		}
	}
	if customResource.Status.Status == v1alpha1.Creating {
		if err := r.installCluster(ctx, &customResource); err != nil {
			return ctrl.Result{}, err
		}
	} else if customResource.Status.Version != customResource.Spec.Version {
		if err := r.patchCluster(ctx, &customResource); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		// check custom resources status
		return r.checkResourceStatus(ctx, &customResource)
	}
	klog.V(1).Info("resource name: ", customResource.Name, ", state: ", customResource.Status.Status)
	return ctrl.Result{}, nil
}

func (r *ManifestReconciler) patchCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return err
	}
	err = r.Client.Patch(ctx, obj, client.Merge)
	if err != nil {
		klog.V(1).Info(err.Error())
		return err
	}

	resource.Status.Version = resource.Spec.Version
	err = r.Client.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.Status = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}

	return nil
}

func (r *ManifestReconciler) deleteCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	klog.V(1).Infof("do delete cluster: %s, %s, %s", resource.Namespace, resource.Name, resource.Spec.Kind)

	obj, err := getUnstructuredObj(resource)
	if err != nil {
		klog.Errorf("get unstructured object error: %s", err.Error())
		return err
	}
	err = r.Delete(ctx, obj)
	return err
}

func (r *ManifestReconciler) installCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	klog.V(1).Infof("install cluster: %s, %s, %s", resource.Namespace, resource.Name, resource.Spec.Kind)
	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return err
	}

	err = r.Create(ctx, obj)
	if err != nil {
		return err
	}

	time.Sleep(500 * time.Millisecond)

	err = r.Get(ctx, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)
	if err != nil {
		klog.Error(err, "get unstructured object error.")
		return err
	}

	clusterStatus := getUnstructuredObjStatus(obj)
	resource.Status.Status = clusterStatus
	resource.Status.Version = resource.Spec.Version
	switch resource.Kind {
	case v1alpha1.DBTypeClickHouse:
		resource.Spec.AppName = v1alpha1.ClusterAppTypeClickHouse
	case v1alpha1.DBTypePostgreSQL:
		resource.Spec.AppName = v1alpha1.ClusterAPPTypePostgreSQL
	case v1alpha1.DBTypeMysql:
		resource.Spec.AppName = v1alpha1.ClusterAPPTypeMySQL
	default:
		resource.Spec.AppName = ""
	}
	err = r.Client.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.Status = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}
	return nil
}

func (r *ManifestReconciler) checkResourceStatus(ctx context.Context, resource *v1alpha1.Manifest) (ctrl.Result, error) {
	klog.V(1).Infof("do check status: %s, %s, %s", resource.Namespace, resource.Name, resource.Spec.Kind)
	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.Get(ctx, types.NamespacedName{
		Namespace: resource.Namespace,
		Name:      resource.Name}, obj)
	if err != nil {
		klog.V(1).Info(err.Error())
	}

	clusterStatus := getUnstructuredObjStatus(obj)

	resource.Status.Status = clusterStatus
	err = r.Client.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.Status = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}
	return ctrl.Result{RequeueAfter: CheckTime}, err
}

func getUnstructuredObj(resource *v1alpha1.Manifest) (obj *unstructured.Unstructured, err error) {
	obj = &unstructured.Unstructured{}
	_, _, err = decUnstructured.Decode([]byte(resource.Spec.CustomResource), nil, obj)
	if err != nil {
		klog.Errorf("decode unstructured object error: %s", err.Error())
	}
	obj.SetName(resource.Name)
	obj.SetNamespace(resource.Namespace)
	return
}

func getUnstructuredObjStatus(obj *unstructured.Unstructured) string {
	var clusterStatus string
	statusMap, ok := obj.Object["status"].(map[string]interface{})
	if ok {
		clusterStatus = statusMap["status"].(string)
	} else {
		clusterStatus = v1alpha1.ClusterStatusUnknown
	}
	return clusterStatus
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManifestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Manifest{}).
		Complete(r)
}

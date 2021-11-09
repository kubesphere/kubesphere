/*
Copyright 2021 The KubeSphere Authors.

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
	"encoding/base64"
	"encoding/json"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/utils/clusterclient"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const CheckTime = 30 * time.Second

var (
	decUnstructured = yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	finalizer       = "finalizers.radondb.com/customresource"
)

// ManifestReconciler reconciles a Manifest object
type ManifestReconciler struct {
	client.Client
	KsFactory          externalversions.SharedInformerFactory
	MultiClusterEnable bool
	clusterClients     clusterclient.ClusterClients
	Scheme             *runtime.Scheme
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

	customResource := &v1alpha1.Manifest{}
	if err := r.Get(ctx, req.NamespacedName, customResource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if customResource.Status.State == "" {
		customResource.Status.State = v1alpha1.ManifestCreating
		err := r.Status().Update(ctx, customResource)
		return reconcile.Result{}, err
	}

	if customResource.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted
		if !sliceutil.HasString(customResource.ObjectMeta.Finalizers, finalizer) {
			clusterName := customResource.GetManifestCluster()
			if r.MultiClusterEnable && clusterName != "" {
				clusterInfo, err := r.clusterClients.Get(clusterName)
				if err != nil {
					// cluster not exists, delete the crd
					klog.Warningf("cluster %s not found, delete the custom resource %s/%s",
						clusterName, customResource.GetManifestNamespace(), customResource.GetName())
					return reconcile.Result{}, r.Delete(context.TODO(), customResource)
				}

				// Host cluster will self-healing, delete host cluster won't cause deletion of  helm release
				if !r.clusterClients.IsHostCluster(clusterInfo) {
					// add owner References
					customResource.OwnerReferences = append(customResource.OwnerReferences, metav1.OwnerReference{
						APIVersion: clusterv1alpha1.SchemeGroupVersion.String(),
						Kind:       clusterv1alpha1.ResourceKindCluster,
						Name:       clusterInfo.Name,
						UID:        clusterInfo.UID,
					})
				}
			}
			customResource.ObjectMeta.Finalizers = append(customResource.ObjectMeta.Finalizers, finalizer)
			err := r.Update(ctx, customResource)
			return reconcile.Result{}, err
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(customResource.ObjectMeta.Finalizers, finalizer) {
			err := r.deleteCluster(ctx, customResource)
			if err != nil {
				klog.Errorf("delete database cluster error: %s", client.IgnoreNotFound(err).Error())
			}
			customResource.ObjectMeta.Finalizers = sliceutil.RemoveString(customResource.ObjectMeta.Finalizers, func(item string) bool {
				if item == finalizer {
					return true
				}
				return false
			})
			err = r.Update(ctx, customResource)
			return reconcile.Result{}, err
		}
	}
	if customResource.Status.State == v1alpha1.ManifestCreating {
		if err := r.installCluster(ctx, customResource); err != nil {
			return ctrl.Result{}, err
		}
	} else if customResource.Status.Version != customResource.Spec.Version {
		return r.updateCluster(ctx, customResource)
	} else {
		// check custom resources status
		return r.checkResourceStatus(ctx, customResource)
	}
	return ctrl.Result{}, nil
}

func (r *ManifestReconciler) updateCluster(ctx context.Context, resource *v1alpha1.Manifest) (ctrl.Result, error) {

	cli, err := r.newClusterClient(resource.GetManifestCluster())
	if err != nil {
		return ctrl.Result{}, err
	}

	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return ctrl.Result{}, err
	}
	oldObj := obj.DeepCopy()
	err = cli.Get(ctx, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, oldObj)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	oldObj.Object["spec"] = obj.Object["spec"]
	err = cli.Patch(ctx, oldObj, client.Merge)
	if err != nil {
		klog.Errorf("update cluster error: %s", err)
		return ctrl.Result{}, err
	}

	resource.Status.Version = resource.Spec.Version
	err = r.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.ResourceState = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}

	return ctrl.Result{}, nil
}

func (r *ManifestReconciler) deleteCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	if resource.Status.State != v1alpha1.FrontTerminating {
		resource.Status.State = v1alpha1.FrontTerminating
		err := r.Status().Update(context.TODO(), resource)
		if err != nil {
			return err
		}
	}

	cli, err := r.newClusterClient(resource.GetManifestCluster())
	if err != nil {
		return err
	}

	obj, err := getUnstructuredObj(resource)
	if err != nil {
		klog.Errorf("get unstructured object error: %s", err.Error())
		return err
	}

	resourceKind := obj.GetKind()
	if resourceKind == v1alpha1.KindPostgreSQLCluster {
		// delete pg cluster
		pgCluster := obj.DeepCopy()
		pgCluster.SetKind(v1alpha1.KindPgCluster)
		pgCluster.SetAPIVersion(v1alpha1.KindPgClusterVersion)
		err = cli.Delete(ctx, pgCluster)
		if err != nil {
			klog.Errorf("delete pgcluster resource error: %s", err.Error())
		}
	} else if resourceKind == v1alpha1.KindMysqlCluster {
		// delete secret of mysql user
		err := r.createOrDeleteMysqlClusterPasswordSecret(ctx, resource, true)
		if err != nil {
			klog.Errorf("delete mysql password secret error: %s", err)
			resource.Status.ResourceState = v1alpha1.Error
			err = r.Status().Update(ctx, resource)
			return err
		}
	}

	err = cli.Delete(ctx, obj)
	return client.IgnoreNotFound(err)
}

func (r *ManifestReconciler) installCluster(ctx context.Context, resource *v1alpha1.Manifest) error {
	klog.Infof("install cluster: %s, %s, %s", resource.Namespace, resource.Name, resource.Spec.Kind)

	cli, err := r.newClusterClient(resource.GetManifestCluster())
	if err != nil {
		return err
	}

	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return err
	}

	err = cli.Get(ctx, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)
	if !errors.IsNotFound(err) {
		klog.Errorf("get obj error: %s", err)
		return err
	}

	err = cli.Create(ctx, obj)
	if err != nil {
		klog.Errorf("create cluster error: %s, %s, %s", err, obj.GetNamespace(), obj.GetName())
		resource.Status.ResourceState = v1alpha1.Error
		err = cli.Status().Update(ctx, resource)
		return err
	}
	resource.Status.State = v1alpha1.ManifestCreated
	resource.Status.ResourceState = v1alpha1.FrontCreating
	resource.Status.Version = resource.Spec.Version
	err = cli.Status().Update(ctx, resource)
	if err != nil {
		klog.Errorf("update manifest status error: %s", err)
		return err
	}

	// If it is a MySQL Cluster, create a secret to save password
	resourceKind := obj.GetKind()
	if resourceKind == v1alpha1.KindMysqlCluster {
		err = r.createOrDeleteMysqlClusterPasswordSecret(ctx, resource, false)
		if err != nil {
			klog.Errorf("create secret error: %s", err)
			resource.Status.ResourceState = v1alpha1.Error
			err = cli.Status().Update(ctx, resource)
			return err
		}
	} else if resourceKind == v1alpha1.KindPostgreSQLCluster {
		err = r.getPostgresPassword(resource, obj)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ManifestReconciler) createOrDeleteMysqlClusterPasswordSecret(ctx context.Context, resource *v1alpha1.Manifest, delete bool) (err error) {
	cli, err := r.newClusterClient(resource.GetManifestCluster())
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.Name + v1alpha1.SuffixSecretName,
			Namespace: resource.Namespace,
		},
	}
	if delete {
		err = cli.Delete(ctx, secret)
	} else {
		err = cli.Create(ctx, secret)
	}
	return
}

func (r *ManifestReconciler) checkResourceStatus(ctx context.Context, resource *v1alpha1.Manifest) (ctrl.Result, error) {
	klog.V(1).Infof("do check status: %s, %s, %s", resource.Namespace, resource.Name, resource.Spec.Kind)
	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	err = r.Get(ctx, types.NamespacedName{
		Namespace: resource.Namespace,
		Name:      resource.Name}, obj)
	if err != nil {
		klog.V(1).Info(err.Error())
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var clusterStatus string
	resourceKind := obj.GetKind()
	if resourceKind == v1alpha1.KindPostgreSQLCluster {
		clusterStatus, _ = r.getPgClusterStatus(ctx, resource)
	} else {
		clusterStatus = getUnstructuredObjStatus(obj)
	}

	resource.Status.ResourceState = clusterStatus
	err = r.Status().Update(ctx, resource)
	if err != nil {
		resource.Status.ResourceState = v1alpha1.Failed
		err = r.Status().Update(ctx, resource)
	}
	return ctrl.Result{RequeueAfter: CheckTime}, err
}

func (r *ManifestReconciler) getPgClusterStatus(ctx context.Context, resource *v1alpha1.Manifest) (string, error) {
	cli, err := r.newClusterClient(resource.GetManifestCluster())
	if err != nil {
		return v1alpha1.Error, err
	}
	obj, err := getUnstructuredObj(resource)
	if err != nil {
		return v1alpha1.Error, err
	}
	obj.SetKind(v1alpha1.KindPgCluster)
	obj.SetAPIVersion(v1alpha1.KindPgClusterVersion)
	err = cli.Get(ctx, types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)
	if err != nil {
		klog.Errorf("get Pgcluster resource error: %s, %s, %s", err, obj.GetNamespace(), obj.GetName())
		if errors.IsNotFound(err) {
			_ = cli.Delete(ctx, obj)
		}
		return v1alpha1.Error, client.IgnoreNotFound(err)
	}
	return getUnstructuredObjStatus(obj), nil
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
	statusMap, ok := obj.Object["status"].(map[string]interface{})
	if !ok {
		return v1alpha1.ClusterStatusUnknown
	}

	clusterStatus, ok := statusMap["state"].(string)
	if !ok {
		return v1alpha1.ClusterStatusUnknown
	}
	return convertObjState(clusterStatus)
}

func (r *ManifestReconciler) getPostgresPassword(manifest *v1alpha1.Manifest, obj *unstructured.Unstructured) error {
	time.Sleep(500 * time.Millisecond)
	secret := &corev1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: manifest.Name + "-postgres-secret", Namespace: manifest.Namespace},
	}
	if err := r.Get(context.TODO(), types.NamespacedName{
		Namespace: manifest.Namespace,
		Name:      manifest.Name + "-postgres-secret",
	}, secret); err != nil {
		klog.Errorf("get postgres user's password error: %s", err)
	}

	pwd, err := base64.StdEncoding.DecodeString(base64.StdEncoding.EncodeToString(secret.Data["password"]))
	if err != nil {
		klog.Errorf("decode base64 string error: %s", err)
	}

	postgres := make(map[string]string)
	postgres["username"] = string(secret.Data["username"])
	postgres["password"] = string(pwd)

	spec, ok := obj.Object["spec"]
	if ok {
		specMap, ok := spec.(map[string]interface{})
		if ok {
			specMap["users"] = []map[string]string{postgres}
		}
	}

	err = r.Get(context.TODO(), types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, obj)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	obj.Object["spec"] = spec
	err = r.Patch(context.TODO(), obj, client.Merge)
	if err != nil {
		klog.Errorf("patch postgresqlcluster resource error: %s", err)
	}

	objBytes, err := json.Marshal(obj)
	if err != nil {
		klog.Errorf("unmarshal unstructured obj error: %s", err)
	}

	err = r.Get(context.TODO(), types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, manifest)
	if err != nil {
		return client.IgnoreNotFound(err)
	}
	manifest.Spec.CustomResource = string(objBytes)
	err = r.Patch(context.TODO(), manifest, client.Merge)
	if err != nil {
		klog.Errorf("patch manifest resource error: %s", err)
	}

	return err
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManifestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	if r.KsFactory != nil && r.MultiClusterEnable {
		r.clusterClients = clusterclient.NewClusterClient(r.KsFactory.Cluster().V1alpha1().Clusters())
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Manifest{}).
		Complete(r)
}

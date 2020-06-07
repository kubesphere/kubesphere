/*
Copyright 2019 The KubeSphere Authors.

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

package namespace

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	"openpitrix.io/openpitrix/pkg/pb"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Namespace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager, openpitrixClient openpitrix.Client) error {
	return add(mgr, newReconciler(mgr, openpitrixClient))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, openpitrixClient openpitrix.Client) reconcile.Reconciler {
	return &ReconcileNamespace{
		Client:           mgr.GetClient(),
		scheme:           mgr.GetScheme(),
		openpitrixClient: openpitrixClient,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespace-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNamespace{}

// ReconcileNamespace reconciles a Namespace object
type ReconcileNamespace struct {
	client.Client
	openpitrixClient openpitrix.Client
	scheme           *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Namespace object and makes changes based on the state read
// and what is in the Namespace.Spec
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.kubesphere.io,resources=namespaces/status,verbs=get;update;patch
func (r *ReconcileNamespace) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Namespace instance
	instance := &corev1.Namespace{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			// The object is being deleted
			// our finalizer is present, so lets handle our external dependency
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// name of your custom finalizer
	finalizer := "finalizers.kubesphere.io/namespaces"

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, finalizer)
			if instance.Labels == nil {
				instance.Labels = make(map[string]string)
			}
			instance.Labels[constants.NamespaceLabelKey] = instance.Name
			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(instance.ObjectMeta.Finalizers, finalizer) {
			if err = r.deleteRouter(instance.Name); err != nil {
				return reconcile.Result{}, err
			}

			// delete runtime
			if err = r.deleteRuntime(instance); err != nil {
				return reconcile.Result{}, err
			}

			// remove our finalizer from the list and update it.
			instance.ObjectMeta.Finalizers = sliceutil.RemoveString(instance.ObjectMeta.Finalizers, func(item string) bool {
				return item == finalizer
			})

			if err := r.Update(context.Background(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Our finalizer has finished, so the reconciler can do nothing.
		return reconcile.Result{}, nil
	}

	if err = r.checkAndBindWorkspace(instance); err != nil {
		return reconcile.Result{}, err
	}

	// skip if openpitrix is not enabled
	if r.openpitrixClient != nil {
		if err := r.checkAndCreateRuntime(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileNamespace) isControlledByWorkspace(namespace *corev1.Namespace) (bool, error) {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	// without workspace label
	if workspaceName == "" {
		return false, nil
	}

	return true, nil
}

// Create openpitrix runtime
func (r *ReconcileNamespace) checkAndCreateRuntime(namespace *corev1.Namespace) error {

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		return nil
	}

	adminKubeConfigName := fmt.Sprintf("kubeconfig-%s", constants.AdminUserName)

	runtimeCredentials, err := r.openpitrixClient.DescribeRuntimeCredentials(openpitrix.SystemContext(), &pb.DescribeRuntimeCredentialsRequest{SearchWord: &wrappers.StringValue{Value: adminKubeConfigName}, Limit: 1})

	if err != nil {
		klog.Error(fmt.Sprintf("create runtime, namespace: %s, error: %s", namespace.Name, err))
		return err
	}

	var kubesphereRuntimeCredentialId string

	// runtime credential exist
	if len(runtimeCredentials.GetRuntimeCredentialSet()) > 0 {
		kubesphereRuntimeCredentialId = runtimeCredentials.GetRuntimeCredentialSet()[0].GetRuntimeCredentialId().GetValue()
	} else {
		adminKubeConfig := corev1.ConfigMap{}
		err := r.Get(context.TODO(), types.NamespacedName{Namespace: constants.KubeSphereControlNamespace, Name: adminKubeConfigName}, &adminKubeConfig)

		if err != nil {
			klog.Error(fmt.Sprintf("create runtime, namespace: %s, error: %s", namespace.Name, err))
			return err
		}

		resp, err := r.openpitrixClient.CreateRuntimeCredential(openpitrix.SystemContext(), &pb.CreateRuntimeCredentialRequest{
			Name:                     &wrappers.StringValue{Value: adminKubeConfigName},
			Provider:                 &wrappers.StringValue{Value: "kubernetes"},
			Description:              &wrappers.StringValue{Value: "kubeconfig"},
			RuntimeUrl:               &wrappers.StringValue{Value: "kubesphere"},
			RuntimeCredentialContent: &wrappers.StringValue{Value: adminKubeConfig.Data["config"]},
		})

		if err != nil {
			klog.Error(fmt.Sprintf("create runtime, namespace: %s, error: %s", namespace.Name, err))
			return err
		}

		kubesphereRuntimeCredentialId = resp.GetRuntimeCredentialId().GetValue()
	}

	// TODO runtime id is invalid when recreate runtime
	runtimeId, err := r.openpitrixClient.CreateRuntime(openpitrix.SystemContext(), &pb.CreateRuntimeRequest{
		Name:                &wrappers.StringValue{Value: namespace.Name},
		RuntimeCredentialId: &wrappers.StringValue{Value: kubesphereRuntimeCredentialId},
		Provider:            &wrappers.StringValue{Value: openpitrix.KubernetesProvider},
		Zone:                &wrappers.StringValue{Value: namespace.Name},
	})

	if err != nil {
		klog.Error(fmt.Sprintf("create runtime, namespace: %s, error: %s", namespace.Name, err))
		return err
	}

	klog.V(4).Infof("runtime created successfully, namespace: %s, runtime id: %s", namespace.Name, runtimeId)

	return nil
}

// Delete openpitrix runtime
func (r *ReconcileNamespace) deleteRuntime(namespace *corev1.Namespace) error {

	if runtimeId := namespace.Annotations[constants.OpenPitrixRuntimeAnnotationKey]; runtimeId != "" {
		_, err := r.openpitrixClient.DeleteRuntimes(openpitrix.SystemContext(), &pb.DeleteRuntimesRequest{RuntimeId: []string{runtimeId}, Force: &wrappers.BoolValue{Value: true}})

		if err == nil || openpitrix.IsNotFound(err) || openpitrix.IsDeleted(err) {
			return nil
		} else {
			klog.Errorf("delete openpitrix runtime: %s, error: %s", runtimeId, err)
			return err
		}
	}

	return nil
}

// Create openpitrix runtime
func (r *ReconcileNamespace) checkAndBindWorkspace(namespace *corev1.Namespace) error {

	workspaceName := namespace.Labels[constants.WorkspaceLabelKey]

	if workspaceName == "" {
		return nil
	}

	workspace := &v1alpha1.Workspace{}

	err := r.Get(context.TODO(), types.NamespacedName{Name: workspaceName}, workspace)

	if err != nil {
		// skip if workspace not found
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
		return err
	}

	if !metav1.IsControlledBy(namespace, workspace) {
		if err := controllerutil.SetControllerReference(workspace, namespace, r.scheme); err != nil {
			klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
			return err
		}
		err = r.Update(context.TODO(), namespace)
		if err != nil {
			klog.Errorf("bind workspace namespace: %s, workspace: %s, error: %s", namespace.Name, workspaceName, err)
			return err
		}
	}

	return nil
}

func (r *ReconcileNamespace) deleteRouter(namespace string) error {
	routerName := constants.IngressControllerPrefix + namespace

	// delete service first
	found := corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
	}

	err = r.Delete(context.TODO(), &found)
	if err != nil {
		klog.Error(err)
		return err
	}

	// delete deployment
	deploy := appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: constants.IngressControllerNamespace, Name: routerName}, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		klog.Error(err)
		return err
	}

	err = r.Delete(context.TODO(), &deploy)
	if err != nil {
		klog.Error(err)
		return err
	}

	return nil

}

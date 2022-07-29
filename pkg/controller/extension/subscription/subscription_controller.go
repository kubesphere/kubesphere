/*
Copyright 2022 KubeSphere Authors

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

package subscription

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/api/application/v1alpha1"
	extensionsv1alpha1 "kubesphere.io/api/extensions/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/controller/extension/util"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmrepoindex"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix/helmwrapper"
)

const (
	SubscriptionFinalizer = "subscription.extensions.kubesphere.io"
)

var _ reconcile.Reconciler = &SubscriptionReconciler{}

type SubscriptionReconciler struct {
	client.Client
}

// reconcileDelete delete the helm release involved and remove finalizer from subscription.
func (r *SubscriptionReconciler) reconcileDelete(ctx context.Context, sub *extensionsv1alpha1.Subscription) (ctrl.Result, error) {
	wrapper := helmwrapper.NewHelmWrapper("", sub.Spec.TargetNamespace, sub.Spec.ReleaseName)

	// TODO: Refactor with helm controller or helm client
	_, err := wrapper.Manifest()
	if err != nil {
		s := err.Error()
		if !strings.Contains(s, "release: not found") {
			return ctrl.Result{}, err
		}
		// The involved release does not exist, just move on.
	} else {
		if err := wrapper.Uninstall(); err != nil {
			klog.Errorf("delete helm release %s/%s failed, error: %s", sub.Spec.TargetNamespace, sub.Spec.ReleaseName, err)
			return ctrl.Result{}, err
		} else {
			klog.Infof("delete helm release %s/%s", sub.Spec.TargetNamespace, sub.Spec.ReleaseName)
		}
	}

	klog.V(4).Infof("remove the finalizer for subscription %s", sub.Name)
	// Remove the finalizer from the subscription and update it.
	controllerutil.RemoveFinalizer(sub, SubscriptionFinalizer)
	if err := r.Update(ctx, sub); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) loadChartData(ctx context.Context, pluginRef *extensionsv1alpha1.PluginRef) (string, error) {
	pluginVer := &extensionsv1alpha1.PluginVersion{}
	err := r.Get(ctx, types.NamespacedName{Name: fmt.Sprintf("%s-%s", pluginRef.Name, pluginRef.Version)}, pluginVer)
	if err != nil {
		return "", err
	}
	repo := &extensionsv1alpha1.Repository{}
	err = r.Get(ctx, types.NamespacedName{Name: pluginVer.Spec.Repo}, repo)
	if err != nil {
		return "", err
	}
	po := &corev1.Pod{}
	podName := util.GeneratePodName(repo.Name)
	if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: podName}, po); err != nil {
		return "", err
	}

	var url string
	for _, d := range pluginVer.Spec.URLs {
		d = strings.TrimPrefix(d, "/")
		if len(d) > 0 {
			url = d
			break
		}
	}
	if len(url) == 0 {
		return "", fmt.Errorf("empty url")
	}

	// TODO: Fetch load data from repo service.
	if po.Status.Phase == corev1.PodRunning {
		buf, err := helmrepoindex.LoadChart(ctx, fmt.Sprintf("http://%s:8080/%s", po.Status.PodIP, url), &v1alpha1.HelmRepoCredential{})
		if err != nil {
			return "", err
		} else {
			return buf.String(), nil
		}
	} else {
		return "", fmt.Errorf("repo not ready")
	}
}

func (r *SubscriptionReconciler) doReconcile(ctx context.Context, sub *extensionsv1alpha1.Subscription) (*extensionsv1alpha1.Subscription, ctrl.Result, error) {
	wrapper := helmwrapper.NewHelmWrapper("", sub.Spec.TargetNamespace, sub.Spec.ReleaseName)
	// TODO: Refactor with helm controller or helm client
	_, err := wrapper.Manifest()
	if err != nil {
		s := err.Error()
		if !strings.Contains(s, "release: not found") {
			return sub, ctrl.Result{}, err
		} else {
			charData, err := r.loadChartData(ctx, &sub.Spec.Plugin)
			if err == nil {
				if err := wrapper.Install(sub.Spec.Plugin.Name, charData, string(sub.Spec.Config)); err != nil {
					klog.Errorf("install helm release %s/%s failed, error: %s", sub.Spec.TargetNamespace, sub.Spec.ReleaseName, err)
					return sub, ctrl.Result{}, err
				} else {
					klog.Infof("install helm release %s/%s", sub.Spec.TargetNamespace, sub.Spec.ReleaseName)
				}
			} else {
				klog.Errorf("fail to load chart data for subscription: %s, error: %s", sub.Name, err)
				return nil, ctrl.Result{}, err
			}
		}
	} else { //nolint:staticcheck
		// TODO: Upgrade the release.
	}

	// TODO: Add more conditions
	sub.Status.State = "installed"
	if err := r.Update(ctx, sub); err != nil {
		return sub, ctrl.Result{}, err
	}

	return sub, ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.V(4).Infof("sync subscription: %s ", req.String())

	sub := &extensionsv1alpha1.Subscription{}
	if err := r.Client.Get(ctx, req.NamespacedName, sub); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !controllerutil.ContainsFinalizer(sub, SubscriptionFinalizer) {
		patch := client.MergeFrom(sub.DeepCopy())
		controllerutil.AddFinalizer(sub, SubscriptionFinalizer)
		if err := r.Patch(ctx, sub, patch); err != nil {
			klog.Errorf("unable to register finalizer for subscription %s, error: %s", sub.Name, err)
			return ctrl.Result{}, err
		}
	}

	if sub.ObjectMeta.DeletionTimestamp != nil {
		return r.reconcileDelete(ctx, sub)
	}

	if _, res, err := r.doReconcile(ctx, sub); err != nil {
		return res, err
	}

	return ctrl.Result{}, nil
}

func (r *SubscriptionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		Named("subscription-controller").
		For(&extensionsv1alpha1.Subscription{}).Complete(r)
}

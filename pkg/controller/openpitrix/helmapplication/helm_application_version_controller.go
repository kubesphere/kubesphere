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

package helmapplication

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver/v3"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/apis/application/v1alpha1"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/sliceutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

const (
	HelmAppVersionFinalizer = "helmappversion.application.kubesphere.io"
)

var _ reconcile.Reconciler = &ReconcileHelmApplicationVersion{}

// ReconcileHelmApplicationVersion reconciles a helm application version object
type ReconcileHelmApplicationVersion struct {
	client.Client
}

// Reconcile reads that state of the cluster for a helmapplicationversions object and makes changes based on the state read
// and what is in the helmapplicationversions.Spec
func (r *ReconcileHelmApplicationVersion) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	start := time.Now()
	klog.V(4).Infof("sync helm application version: %s", request.String())
	defer func() {
		klog.V(4).Infof("sync helm application version end: %s, elapsed: %v", request.String(), time.Now().Sub(start))
	}()

	appVersion := &v1alpha1.HelmApplicationVersion{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, appVersion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if appVersion.ObjectMeta.DeletionTimestamp.IsZero() {

		if appVersion.Status.State == "" {
			// set status to draft
			return reconcile.Result{}, r.updateStatus(appVersion)
		}

		if !sliceutil.HasString(appVersion.ObjectMeta.Finalizers, HelmAppVersionFinalizer) {
			appVersion.ObjectMeta.Finalizers = append(appVersion.ObjectMeta.Finalizers, HelmAppVersionFinalizer)
			if err := r.Update(context.Background(), appVersion); err != nil {
				return reconcile.Result{}, err
			} else {
				return reconcile.Result{}, nil
			}
		}
	} else {
		// The object is being deleted
		if sliceutil.HasString(appVersion.ObjectMeta.Finalizers, HelmAppVersionFinalizer) {
			// update related helm application
			err = updateHelmApplicationStatus(r.Client, appVersion.GetHelmApplicationId(), false)
			if err != nil {
				return reconcile.Result{}, err
			}

			err = updateHelmApplicationStatus(r.Client, appVersion.GetHelmApplicationId(), true)
			if err != nil {
				return reconcile.Result{}, err
			}

			// Delete HelmApplicationVersion
			appVersion.ObjectMeta.Finalizers = sliceutil.RemoveString(appVersion.ObjectMeta.Finalizers, func(item string) bool {
				if item == HelmAppVersionFinalizer {
					return true
				}
				return false
			})
			if err := r.Update(context.Background(), appVersion); err != nil {
				return reconcile.Result{}, err
			}
		}
		return reconcile.Result{}, nil
	}

	// update related helm application
	err = updateHelmApplicationStatus(r.Client, appVersion.GetHelmApplicationId(), false)
	if err != nil {
		return reconcile.Result{}, err
	}

	if appVersion.Status.State == v1alpha1.StateActive {
		// add labels to helm application version
		// The label will exists forever, since this helmapplicationversion's state only can be active and suspend.
		if appVersion.GetHelmRepoId() == "" {
			instanceCopy := appVersion.DeepCopy()
			instanceCopy.Labels[constants.ChartRepoIdLabelKey] = v1alpha1.AppStoreRepoId
			patch := client.MergeFrom(appVersion)
			err = r.Client.Patch(context.TODO(), instanceCopy, patch)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		app := v1alpha1.HelmApplication{}
		err = r.Get(context.TODO(), types.NamespacedName{Name: appVersion.GetHelmApplicationId()}, &app)
		if err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, updateHelmApplicationStatus(r.Client, appVersion.GetHelmApplicationId(), true)
	} else if appVersion.Status.State == v1alpha1.StateSuspended {
		return reconcile.Result{}, updateHelmApplicationStatus(r.Client, appVersion.GetHelmApplicationId(), true)
	}

	return reconcile.Result{}, nil
}

func updateHelmApplicationStatus(c client.Client, appId string, inAppStore bool) error {
	app := v1alpha1.HelmApplication{}

	var err error
	if inAppStore {
		// application name ends with `-store`
		err = c.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s%s", appId, v1alpha1.HelmApplicationAppStoreSuffix)}, &app)
	} else {
		err = c.Get(context.TODO(), types.NamespacedName{Name: appId}, &app)
	}

	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if !app.DeletionTimestamp.IsZero() {
		return nil
	}

	var versions v1alpha1.HelmApplicationVersionList
	err = c.List(context.TODO(), &versions, client.MatchingLabels{
		constants.ChartApplicationIdLabelKey: appId,
	})

	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	latestVersionName := getLatestVersionName(versions, inAppStore)
	state := mergeApplicationVersionState(versions)

	now := time.Now()
	if state != app.Status.State {
		// update StatusTime when state changed
		app.Status.StatusTime = &metav1.Time{Time: now}
	}

	if state != app.Status.State || latestVersionName != app.Status.LatestVersion {
		app.Status.State = state
		app.Status.LatestVersion = latestVersionName
		app.Status.UpdateTime = &metav1.Time{Time: now}
		err := c.Status().Update(context.TODO(), &app)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileHelmApplicationVersion) updateStatus(appVersion *v1alpha1.HelmApplicationVersion) error {
	appVersion.Status = v1alpha1.HelmApplicationVersionStatus{
		State: v1alpha1.StateDraft,
		Audit: []v1alpha1.Audit{
			{
				State:    v1alpha1.StateDraft,
				Time:     appVersion.CreationTimestamp,
				Operator: appVersion.GetCreator(),
			},
		},
	}

	err := r.Status().Update(context.TODO(), appVersion)
	if err != nil {
		return err
	}
	return nil
}

// getLatestVersionName get the latest version of versions.
// if inAppStore is false, get the latest version name of all of the versions
// if inAppStore is true, get the latest version name of the ACTIVE versions.
func getLatestVersionName(versions v1alpha1.HelmApplicationVersionList, inAppStore bool) string {
	if len(versions.Items) == 0 {
		return ""
	}

	var latestVersionName string
	var latestSemver *semver.Version

	for _, version := range versions.Items {
		// If the appVersion is being deleted, ignore it.
		// If inAppStore is true, we just need ACTIVE appVersion.
		if version.DeletionTimestamp != nil || (inAppStore && version.Status.State != v1alpha1.StateActive) {
			continue
		}

		currSemver, err := semver.NewVersion(version.GetSemver())
		if err == nil {
			if latestSemver == nil {
				// the first valid semver
				latestSemver = currSemver
				latestVersionName = version.GetVersionName()
			} else if latestSemver.LessThan(currSemver) {
				// find a newer valid semver
				latestSemver = currSemver
				latestVersionName = version.GetVersionName()
			}
		} else {
			// If the semver is invalid, just ignore it.
			klog.V(2).Infof("parse version failed, id: %s, err: %s", version.Name, err)
		}
	}

	return latestVersionName
}

func mergeApplicationVersionState(versions v1alpha1.HelmApplicationVersionList) string {
	states := make(map[string]int, len(versions.Items))

	for _, version := range versions.Items {
		if version.DeletionTimestamp == nil {
			state := version.Status.State
			states[state] = states[state] + 1
		}
	}

	// If there is one or more active appVersion, the helm application is active
	if states[v1alpha1.StateActive] > 0 {
		return v1alpha1.StateActive
	}

	// All appVersion is draft, the helm application is draft
	if states[v1alpha1.StateDraft] == len(versions.Items) {
		return v1alpha1.StateDraft
	}

	// No active appVersion or draft appVersion, then the app state is suspended
	if states[v1alpha1.StateSuspended] > 0 {
		return v1alpha1.StateSuspended
	}

	// default state is draft
	return v1alpha1.StateDraft
}

func (r *ReconcileHelmApplicationVersion) SetupWithManager(mgr ctrl.Manager) error {
	r.Client = mgr.GetClient()
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.HelmApplicationVersion{}).
		Complete(r)
}

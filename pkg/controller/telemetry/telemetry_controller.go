/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package telemetry

import (
	"context"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	ControllerName = "telemetry"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

func (r *Reconciler) Name() string {
	return ControllerName
}

func (r *Reconciler) Enabled(clusterRole string) bool {
	return strings.EqualFold(clusterRole, string(clusterv1alpha1.ClusterRoleHost))
}

func (r *Reconciler) Hidden() bool {
	return true
}

type Reconciler struct {
	*TelemetryOptions
	runtimeclient.Client
	telemetryRunnable *runnable
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	return builder.
		ControllerManagedBy(mgr).
		For(&corev1.Secret{}, builder.WithPredicates(predicate.NewPredicateFuncs(func(obj runtimeclient.Object) bool {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				return false
			}
			return secret.Namespace == constants.KubeSphereNamespace &&
				secret.Name == ConfigName &&
				secret.Type == constants.SecretTypeGenericPlatformConfig
		}))).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ConfigName,
			Namespace: constants.KubeSphereNamespace,
		},
	}
	if err := r.Client.Get(ctx, runtimeclient.ObjectKeyFromObject(secret), secret); err != nil {
		if errors.IsNotFound(err) {
			// not found. telemetry is disabled.
			if r.telemetryRunnable != nil {
				r.telemetryRunnable.Close()
			}
			return reconcile.Result{}, nil
		}
		klog.V(9).ErrorS(err, "cannot get telemetry option secret")
		return reconcile.Result{RequeueAfter: time.Second}, nil
	}
	// ignore delete resource
	if secret.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	// get config from configmap
	conf, err := LoadTelemetryConfig(secret)
	if err != nil {
		klog.V(9).ErrorS(err, "cannot log telemetry option secret")
		return reconcile.Result{}, nil
	}

	// check value when telemetry is enabled.
	if conf.Enabled && (conf.Endpoint == "" || conf.Schedule == "") {
		klog.V(9).ErrorS(nil, "endpoint and schedule should not be empty when telemetry enabled is true.")
		return reconcile.Result{}, nil
	}

	// stop telemetryRunnable when telemetry is disabled.
	if !conf.Enabled {
		if r.telemetryRunnable != nil {
			r.telemetryRunnable.Close()
		}
		return reconcile.Result{}, nil
	}

	if r.telemetryRunnable == nil {
		r.TelemetryOptions = conf
		r.telemetryRunnable, err = NewRunnable(ctx, r.TelemetryOptions, r.Client)
		if err != nil {
			klog.V(9).ErrorS(err, "failed to new telemetryRunnable")
		}
		return reconcile.Result{}, nil
	}

	if conf.Schedule != r.TelemetryOptions.Schedule {
		r.TelemetryOptions = conf
		if err := r.telemetryRunnable.UpdateSchedule(conf.Schedule); err != nil {
			klog.V(9).ErrorS(err, "failed to update telemetryRunnable schedule")
		}
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

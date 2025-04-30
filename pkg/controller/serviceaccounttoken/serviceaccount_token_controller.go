/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package serviceaccounttoken

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"

	"kubesphere.io/kubesphere/pkg/models/kubeconfig"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	controllerName                 = "service-account-token"
	userKubeConfigSecretNameFormat = "kubeconfig-%s"
	kubeconfigFileName             = "config"
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

type Reconciler struct {
	client.Client
	recorder record.EventRecorder
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	if mgr.KubeconfigOptions.AuthMode != kubeconfig.AuthModeServiceAccountToken {
		klog.Infof("Skip %s controller as the auth mode is not service account token", controllerName)
		return nil
	}

	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()
	return builder.
		ControllerManagedBy(mgr).
		For(&corev1.Secret{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
				if object.GetNamespace() == constants.KubeSphereNamespace && object.GetLabels()[constants.UsernameLabelKey] != "" {
					return true
				}
				return false
			})),
		).
		Named(controllerName).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if username := secret.Labels[constants.UsernameLabelKey]; username != "" {
		if secret.Data == nil {
			return ctrl.Result{}, nil
		}
		token := secret.Data["token"]

		if len(token) > 0 {
			if err := r.UpdateKubeConfigServiceAccountToken(ctx, username, string(token)); err != nil {
				// kubeconfig not generated
				return ctrl.Result{}, err
			}
		}
	}

	r.recorder.Event(secret, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) UpdateKubeConfigServiceAccountToken(ctx context.Context, username string, token string) error {
	secretName := fmt.Sprintf(userKubeConfigSecretNameFormat, username)
	kubeconfigSecret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: secretName}, kubeconfigSecret); err != nil {
		return client.IgnoreNotFound(err)
	}

	kubeconfigSecret = applyToken(kubeconfigSecret, token)

	if err := r.Update(ctx, kubeconfigSecret); err != nil {
		klog.Errorf("Failed to update secret %s: %v", secretName, err)
		return err
	}
	return nil
}

func applyToken(secret *corev1.Secret, token string) *corev1.Secret {
	data := secret.Data[kubeconfigFileName]
	kubeconfig, err := clientcmd.Load(data)
	if err != nil {
		klog.Error(err)
		return secret
	}

	username := secret.Labels[constants.UsernameLabelKey]
	kubeconfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		username: {
			Token: token,
		},
	}

	data, err = clientcmd.Write(*kubeconfig)
	if err != nil {
		return secret
	}

	secret.StringData = map[string]string{kubeconfigFileName: string(data)}
	return secret
}

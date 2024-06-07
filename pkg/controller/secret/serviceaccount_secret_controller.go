/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package secret

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/tools/record"
	corev1alpha1 "kubesphere.io/api/core/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/apiserver/authentication/oauth"
	"kubesphere.io/kubesphere/pkg/apiserver/authentication/token"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	serviceAccountSecretController = "serviceaccount-secret"
	serviceAccountUsernameFormat   = corev1alpha1.ServiceAccountGroup + ":%s:%s"
)

var _ kscontroller.Controller = &ServiceAccountSecretReconciler{}
var _ reconcile.Reconciler = &ServiceAccountSecretReconciler{}

type ServiceAccountSecretReconciler struct {
	client.Client
	Logger        logr.Logger
	EventRecorder record.EventRecorder
	TokenIssuer   token.Issuer
}

func (r *ServiceAccountSecretReconciler) Name() string {
	return serviceAccountSecretController
}

func (r *ServiceAccountSecretReconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	issuer, err := token.NewIssuer(mgr.AuthenticationOptions.Issuer)
	if err != nil {
		return fmt.Errorf("failed to create token issuer: %v", err)
	}
	r.TokenIssuer = issuer
	r.Client = mgr.GetClient()
	r.EventRecorder = mgr.GetEventRecorderFor(serviceAccountSecretController)
	r.Logger = ctrl.Log.WithName("controllers").WithName(serviceAccountSecretController)
	return builder.
		ControllerManagedBy(mgr).
		For(&v1.Secret{}).
		WithEventFilter(predicate.ResourceVersionChangedPredicate{}).
		Complete(r)
}

func (r *ServiceAccountSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.WithValues(req.NamespacedName, "Secret")
	secret := &v1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if secret.Type != corev1alpha1.SecretTypeServiceAccountToken {
		return ctrl.Result{}, nil
	}

	saName := secret.Annotations[corev1alpha1.ServiceAccountName]
	if secret.Data[corev1alpha1.ServiceAccountToken] == nil &&
		saName != "" {
		sa := &corev1alpha1.ServiceAccount{}

		if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: saName}, sa); err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			}
			logger.Error(err, "get serviceaccount failed")
			return ctrl.Result{}, err
		}

		tokenTo, err := r.issueTokenTo(sa)
		if err != nil {
			logger.Error(err, "issue token failed")
			return ctrl.Result{}, err
		}
		if secret.Data == nil {
			secret.Data = make(map[string][]byte, 0)
		}
		secret.Data[corev1alpha1.ServiceAccountToken] = []byte(tokenTo.AccessToken)
		if err = r.Update(ctx, secret); err != nil {
			logger.Error(err, "update secret failed")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ServiceAccountSecretReconciler) issueTokenTo(sa *corev1alpha1.ServiceAccount) (*oauth.Token, error) {
	// We verify that the token is valid by checking the validity of SA, so we issue a token with no expiration date
	accessToken, err := r.TokenIssuer.IssueTo(&token.IssueRequest{
		User: &user.DefaultInfo{
			Name: fmt.Sprintf(serviceAccountUsernameFormat, sa.Namespace, sa.Name),
		},
		Claims: token.Claims{TokenType: token.StaticToken},
	})
	if err != nil {
		return nil, err
	}

	result := oauth.Token{
		AccessToken: accessToken,
		// The OAuth 2.0 token_type response parameter value MUST be Bearer,
		// as specified in OAuth 2.0 Bearer Token Usage [RFC6750]
		TokenType: "Bearer",
	}
	return &result, nil
}

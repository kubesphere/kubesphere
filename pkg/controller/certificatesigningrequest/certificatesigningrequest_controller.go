/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package certificatesigningrequest

import (
	"context"
	"fmt"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const (
	controllerName                 = "csr"
	userKubeConfigSecretNameFormat = "kubeconfig-%s"
	kubeconfigFileName             = "config"
	privateKeyAnnotation           = "kubesphere.io/private-key"
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
	r.recorder = mgr.GetEventRecorderFor(controllerName)
	r.Client = mgr.GetClient()
	return builder.
		ControllerManagedBy(mgr).
		For(&certificatesv1.CertificateSigningRequest{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
				csr := object.(*certificatesv1.CertificateSigningRequest)
				return csr.Labels[constants.UsernameLabelKey] != ""
			})),
		).
		Named(controllerName).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Get the CertificateSigningRequest with this name
	csr := &certificatesv1.CertificateSigningRequest{}
	if err := r.Get(ctx, req.NamespacedName, csr); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// csr create by kubesphere auto approval
	if username := csr.Labels[constants.UsernameLabelKey]; username != "" {
		if err := r.Approve(csr); err != nil {
			return ctrl.Result{}, err
		}
		// certificate data is not empty
		if len(csr.Status.Certificate) > 0 {
			if err := r.UpdateKubeConfig(ctx, username, csr); err != nil {
				// kubeconfig not generated
				return ctrl.Result{}, err
			}
			// release
			if err := r.Delete(ctx, csr, &client.DeleteOptions{GracePeriodSeconds: ptr.To[int64](0)}); err != nil {
				return ctrl.Result{}, err
			}
		}
	}

	r.recorder.Event(csr, corev1.EventTypeNormal, kscontroller.Synced, kscontroller.MessageResourceSynced)
	return ctrl.Result{}, nil
}

func (r *Reconciler) Approve(csr *certificatesv1.CertificateSigningRequest) error {
	// is approved
	if len(csr.Status.Certificate) > 0 {
		return nil
	}
	csr.Status = certificatesv1.CertificateSigningRequestStatus{
		Conditions: []certificatesv1.CertificateSigningRequestCondition{{
			Status:  corev1.ConditionTrue,
			Type:    "Approved",
			Reason:  "KubeSphereApprove",
			Message: "This CSR was approved by KubeSphere",
			LastUpdateTime: metav1.Time{
				Time: time.Now(),
			},
		}},
	}

	// approve csr
	if err := r.SubResource("approval").Update(context.Background(), csr, &client.SubResourceUpdateOptions{SubResourceBody: csr}); err != nil {
		return err
	}
	return nil
}

func (r *Reconciler) UpdateKubeConfig(ctx context.Context, username string, csr *certificatesv1.CertificateSigningRequest) error {
	secretName := fmt.Sprintf(userKubeConfigSecretNameFormat, username)
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Namespace: constants.KubeSphereNamespace, Name: secretName}, secret); err != nil {
		return client.IgnoreNotFound(err)
	}
	secret = applyCert(secret, csr)
	if err := r.Update(ctx, secret); err != nil {
		klog.Errorf("Failed to update secret %s: %v", secretName, err)
		return err
	}
	return nil
}

func applyCert(secret *corev1.Secret, csr *certificatesv1.CertificateSigningRequest) *corev1.Secret {
	data := secret.Data[kubeconfigFileName]
	kubeconfig, err := clientcmd.Load(data)
	if err != nil {
		klog.Error(err)
		return secret
	}

	username := secret.Labels[constants.UsernameLabelKey]
	privateKey := csr.Annotations[privateKeyAnnotation]
	clientCert := csr.Status.Certificate
	kubeconfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		username: {
			ClientKeyData:         []byte(privateKey),
			ClientCertificateData: clientCert,
		},
	}

	data, err = clientcmd.Write(*kubeconfig)
	if err != nil {
		return secret
	}

	delete(secret.Annotations, "csr")
	secret.StringData = map[string]string{kubeconfigFileName: string(data)}
	return secret
}

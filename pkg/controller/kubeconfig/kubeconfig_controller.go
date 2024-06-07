/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package kubeconfig

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	certificatesv1 "k8s.io/api/certificates/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/utils/pkiutil"

	kscontroller "kubesphere.io/kubesphere/pkg/controller"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	controllerName = "kubeconfig"
	residual       = 30 * 24 * time.Hour
)

var _ kscontroller.Controller = &Reconciler{}
var _ reconcile.Reconciler = &Reconciler{}

// Reconciler reconciles a User object
type Reconciler struct {
	client.Client
	config *rest.Config
}

func (r *Reconciler) Name() string {
	return controllerName
}

func (r *Reconciler) SetupWithManager(mgr *kscontroller.Manager) error {
	r.Client = mgr.GetClient()
	r.config = mgr.K8sClient.Config()
	return ctrl.NewControllerManagedBy(mgr).
		Named(controllerName).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		For(&corev1.Secret{},
			builder.WithPredicates(predicate.NewPredicateFuncs(func(object client.Object) bool {
				secret := object.(*corev1.Secret)
				return secret.Namespace == constants.KubeSphereNamespace &&
					secret.Type == kubeconfig.SecretTypeKubeConfig &&
					secret.Labels[constants.UsernameLabelKey] != ""
			}))).
		Complete(r)
}

func (r *Reconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !secret.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if err := r.UpdateSecret(ctx, secret); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) UpdateSecret(ctx context.Context, secret *corev1.Secret) error {
	// already exist and cert will not expire in 3 days
	if isValid(secret) {
		return nil
	}

	// create a new CSR
	var ca []byte
	var err error
	if len(r.config.CAData) > 0 {
		ca = r.config.CAData
	} else {
		ca, err = os.ReadFile(kubeconfig.InClusterCAFilePath)
		if err != nil {
			klog.Errorf("Failed to read CA file: %v", err)
			return err
		}
	}

	username := secret.Labels[constants.UsernameLabelKey]

	currentContext := fmt.Sprintf("%s@%s", username, kubeconfig.DefaultClusterName)
	config := clientcmdapi.Config{
		Kind:        "Config",
		APIVersion:  "v1",
		Preferences: clientcmdapi.Preferences{},
		Clusters: map[string]*clientcmdapi.Cluster{kubeconfig.DefaultClusterName: {
			Server:                   r.config.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: ca,
		}},
		Contexts: map[string]*clientcmdapi.Context{currentContext: {
			Cluster:   kubeconfig.DefaultClusterName,
			AuthInfo:  username,
			Namespace: kubeconfig.DefaultNamespace,
		}},
		AuthInfos:      make(map[string]*clientcmdapi.AuthInfo),
		CurrentContext: currentContext,
	}

	data, err := clientcmd.Write(config)
	if err != nil {
		klog.Errorf("Failed to write kubeconfig for user %s: %v", username, err)
		return err
	}

	if secret.Annotations == nil {
		secret.Annotations = make(map[string]string)
	}

	secret.Data = map[string][]byte{kubeconfig.FileName: data}
	if err = r.Update(ctx, secret); err != nil {
		klog.Errorf("Failed to update kubeconfig for user %s: %v", username, err)
		return err
	}

	if err = r.createCSR(ctx, username); err != nil {
		klog.Errorf("Failed to create CSR for user %s: %v", username, err)
		return err
	}

	return nil
}

func isValid(secret *corev1.Secret) bool {
	username := secret.Labels[constants.UsernameLabelKey]

	data := secret.Data[kubeconfig.FileName]
	if len(data) == 0 {
		return false
	}

	config, err := clientcmd.Load(data)
	if err != nil {
		klog.Warningf("Failed to load kubeconfig for user %s: %v", username, err)
		return false
	}

	if authInfo, ok := config.AuthInfos[username]; ok {
		clientCert, err := certutil.ParseCertsPEM(authInfo.ClientCertificateData)
		if err != nil {
			klog.Warningf("Failed to parse client certificate for user %s: %v", username, err)
			return false
		}
		for _, cert := range clientCert {
			if cert.NotAfter.After(time.Now().Add(residual)) {
				return true
			}
		}
	} else {
		// in process
		return true
	}

	return false
}

func (r *Reconciler) createCSR(ctx context.Context, username string) error {
	csrConfig := &certutil.Config{
		CommonName:   username,
		Organization: nil,
		AltNames:     certutil.AltNames{},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	x509csr, x509key, err := pkiutil.NewCSRAndKey(csrConfig)
	if err != nil {
		klog.Errorf("Failed to create CSR and key for user %s: %v", username, err)
		return err
	}

	var csrBuffer, keyBuffer bytes.Buffer
	if err = pem.Encode(&keyBuffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(x509key)}); err != nil {
		klog.Errorf("Failed to encode private key for user %s: %v", username, err)
		return err
	}

	var csrBytes []byte
	if csrBytes, err = x509.CreateCertificateRequest(rand.Reader, x509csr, x509key); err != nil {
		klog.Errorf("Failed to create CSR for user %s: %v", username, err)
		return err
	}

	if err = pem.Encode(&csrBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		klog.Errorf("Failed to encode CSR for user %s: %v", username, err)
		return err
	}

	csr := csrBuffer.Bytes()
	key := keyBuffer.Bytes()
	csrName := fmt.Sprintf("%s-csr-%d", username, time.Now().Unix())
	k8sCSR := &certificatesv1.CertificateSigningRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        csrName,
			Labels:      map[string]string{constants.UsernameLabelKey: username},
			Annotations: map[string]string{kubeconfig.PrivateKeyAnnotation: string(key)},
		},
		Spec: certificatesv1.CertificateSigningRequestSpec{
			Request:    csr,
			SignerName: certificatesv1.KubeAPIServerClientSignerName,
			Usages:     []certificatesv1.KeyUsage{certificatesv1.UsageKeyEncipherment, certificatesv1.UsageClientAuth, certificatesv1.UsageDigitalSignature},
			Username:   username,
			Groups:     []string{user.AllAuthenticated},
		},
	}

	if err = r.Create(ctx, k8sCSR); err != nil {
		klog.Errorf("Failed to create CSR for user %s: %v", username, err)
		return err
	}

	return nil
}

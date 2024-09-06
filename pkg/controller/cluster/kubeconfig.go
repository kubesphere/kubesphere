/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package cluster

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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusterv1alpha1 "kubesphere.io/api/cluster/v1alpha1"

	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/models/kubeconfig"
	"kubesphere.io/kubesphere/pkg/utils/pkiutil"
)

func (r *Reconciler) updateKubeConfigExpirationDateCondition(
	ctx context.Context, cluster *clusterv1alpha1.Cluster, clusterClient client.Client, config *rest.Config,
) error {
	// we don't need to check member clusters which using proxy mode, their certs are managed and will be renewed by tower.
	if cluster.Spec.Connection.Type == clusterv1alpha1.ConnectionTypeProxy {
		return nil
	}

	klog.V(4).Infof("sync KubeConfig expiration date for cluster %s", cluster.Name)
	cert, err := parseKubeConfigCert(config)
	if err != nil {
		return fmt.Errorf("parseKubeConfigCert for cluster %s failed: %v", cluster.Name, err)
	}
	if cert == nil || cert.NotAfter.IsZero() {
		// delete the KubeConfigCertExpiresInSevenDays condition if it has
		conditions := make([]clusterv1alpha1.ClusterCondition, 0)
		for _, condition := range cluster.Status.Conditions {
			if condition.Type == clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays {
				continue
			}
			conditions = append(conditions, condition)
		}
		cluster.Status.Conditions = conditions
		return nil
	}
	seconds := time.Until(cert.NotAfter).Seconds()
	if seconds/86400 <= 7 {
		if err = r.renewKubeConfig(ctx, cluster, clusterClient, config, cert); err != nil {
			return err
		}
	}

	r.updateClusterCondition(cluster, clusterv1alpha1.ClusterCondition{
		Type:               clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays,
		LastUpdateTime:     metav1.Now(),
		LastTransitionTime: metav1.Now(),
		Reason:             string(clusterv1alpha1.ClusterKubeConfigCertExpiresInSevenDays),
		Message:            cert.NotAfter.String(),
	})
	return nil
}

func parseKubeConfigCert(config *rest.Config) (*x509.Certificate, error) {
	if config.CertData == nil {
		return nil, nil
	}
	block, _ := pem.Decode(config.CertData)
	if block == nil {
		return nil, fmt.Errorf("pem.Decode failed, got empty block data")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

func (r *Reconciler) renewKubeConfig(
	ctx context.Context, cluster *clusterv1alpha1.Cluster, clusterClient client.Client, config *rest.Config, cert *x509.Certificate,
) error {
	apiConfig, err := clientcmd.Load(cluster.Spec.Connection.KubeConfig)
	if err != nil {
		return err
	}
	currentContext := apiConfig.Contexts[apiConfig.CurrentContext]
	username := currentContext.AuthInfo
	authInfo := apiConfig.AuthInfos[username]
	if authInfo.Token != "" {
		return nil
	}

	for _, v := range cert.Subject.Organization {
		// we cannot update the certificate of the system:masters group and will use the certificate of the admin user directly
		// certificatesigningrequests.certificates.k8s.io is forbidden:
		// use of kubernetes.io/kube-apiserver-client signer with system:masters group is not allowed
		//
		// for cases where we can't issue a certificate, we use the token of the kubesphere service account directly
		if v == user.SystemPrivilegedGroup {
			data, err := setKubeSphereSAToken(ctx, clusterClient, apiConfig, username)
			if err != nil {
				return err
			}
			cluster.Spec.Connection.KubeConfig = data
			return nil
		}
	}

	kubeconfig, err := genKubeConfig(ctx, clusterClient, config, username)
	if err != nil {
		return err
	}
	cluster.Spec.Connection.KubeConfig = kubeconfig
	return nil
}

func setKubeSphereSAToken(
	ctx context.Context, clusterClient client.Client, apiConfig *clientcmdapi.Config, username string,
) ([]byte, error) {
	secrets := &corev1.SecretList{}
	if err := clusterClient.List(ctx, secrets,
		client.InNamespace(constants.KubeSphereNamespace),
		client.MatchingLabels{"kubesphere.io/service-account-token": ""},
	); err != nil {
		return nil, err
	}
	var secret *corev1.Secret
	for i, item := range secrets.Items {
		if item.Type == corev1.SecretTypeServiceAccountToken {
			secret = &secrets.Items[i]
			break
		}
	}
	if secret == nil {
		return nil, fmt.Errorf("no kubesphere ServiceAccount secret found")
	}
	apiConfig.AuthInfos = map[string]*clientcmdapi.AuthInfo{
		username: {
			Token: string(secret.Data["token"]),
		},
	}
	data, err := clientcmd.Write(*apiConfig)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func genKubeConfig(ctx context.Context, clusterClient client.Client, clusterConfig *rest.Config, username string) ([]byte, error) {
	csrName, err := createCSR(ctx, clusterClient, username)
	if err != nil {
		return nil, err
	}

	var privateKey, clientCert []byte
	if err = wait.PollUntilContextTimeout(ctx, time.Second*3, time.Minute, false, func(ctx context.Context) (bool, error) {
		csr := &certificatesv1.CertificateSigningRequest{}
		if err = clusterClient.Get(ctx, types.NamespacedName{Name: csrName}, csr); err != nil {
			return false, err
		}
		if len(csr.Status.Certificate) == 0 {
			return false, nil
		}
		privateKey = []byte(csr.Annotations[kubeconfig.PrivateKeyAnnotation])
		clientCert = csr.Status.Certificate
		return true, nil
	}); err != nil {
		return nil, err
	}

	var ca []byte
	if len(clusterConfig.CAData) > 0 {
		ca = clusterConfig.CAData
	} else {
		ca, err = os.ReadFile(kubeconfig.InClusterCAFilePath)
		if err != nil {
			klog.Errorf("Failed to read CA file: %v", err)
			return nil, err
		}
	}

	currentContext := fmt.Sprintf("%s@%s", username, kubeconfig.DefaultClusterName)
	config := clientcmdapi.Config{
		Kind:        "Config",
		APIVersion:  "v1",
		Preferences: clientcmdapi.Preferences{},
		Clusters: map[string]*clientcmdapi.Cluster{kubeconfig.DefaultClusterName: {
			Server:                   clusterConfig.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: ca,
		}},
		Contexts: map[string]*clientcmdapi.Context{currentContext: {
			Cluster:   kubeconfig.DefaultClusterName,
			AuthInfo:  username,
			Namespace: kubeconfig.DefaultNamespace,
		}},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			username: {
				ClientKeyData:         privateKey,
				ClientCertificateData: clientCert,
			},
		},
		CurrentContext: currentContext,
	}
	return clientcmd.Write(config)
}

func createCSR(ctx context.Context, clusterClient client.Client, username string) (string, error) {
	x509csr, x509key, err := pkiutil.NewCSRAndKey(&certutil.Config{
		CommonName:   username,
		Organization: nil,
		AltNames:     certutil.AltNames{},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	})
	if err != nil {
		klog.Errorf("Failed to create CSR and key for user %s: %v", username, err)
		return "", err
	}

	var csrBuffer, keyBuffer bytes.Buffer
	if err = pem.Encode(&keyBuffer, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(x509key)}); err != nil {
		klog.Errorf("Failed to encode private key for user %s: %v", username, err)
		return "", err
	}

	var csrBytes []byte
	if csrBytes, err = x509.CreateCertificateRequest(rand.Reader, x509csr, x509key); err != nil {
		klog.Errorf("Failed to create CSR for user %s: %v", username, err)
		return "", err
	}

	if err = pem.Encode(&csrBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes}); err != nil {
		klog.Errorf("Failed to encode CSR for user %s: %v", username, err)
		return "", err
	}

	csr := csrBuffer.Bytes()
	key := keyBuffer.Bytes()
	csrName := fmt.Sprintf("%s-csr-%d", username, time.Now().Unix())
	k8sCSR := &certificatesv1.CertificateSigningRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:        csrName,
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

	if err = clusterClient.Create(ctx, k8sCSR); err != nil {
		klog.Errorf("Failed to create CSR for user %s: %v", username, err)
		return "", err
	}
	return approveCSR(ctx, clusterClient, k8sCSR)
}

func approveCSR(ctx context.Context, clusterClient client.Client, csr *certificatesv1.CertificateSigningRequest) (string, error) {
	csr.Status = certificatesv1.CertificateSigningRequestStatus{
		Conditions: []certificatesv1.CertificateSigningRequestCondition{{
			Status:  corev1.ConditionTrue,
			Type:    certificatesv1.CertificateApproved,
			Reason:  "KubeSphereApprove",
			Message: "This CSR was approved by KubeSphere",
			LastUpdateTime: metav1.Time{
				Time: time.Now(),
			},
		}},
	}

	if err := clusterClient.SubResource("approval").Update(ctx, csr, &client.SubResourceUpdateOptions{SubResourceBody: csr}); err != nil {
		return "", err
	}
	return csr.Name, nil
}

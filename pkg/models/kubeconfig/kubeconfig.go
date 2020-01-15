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

package kubeconfig

import (
	"bytes"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/constants"
	pkiutil "kubesphere.io/kubesphere/pkg/models/kubeconfig/internal"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"time"
)

const (
	kubeconfigNameFormat = "kubeconfig-%s"
	defaultClusterName   = "local"
	defaultNamespace     = "default"
	fileName             = "config"
	configMapKind        = "ConfigMap"
	configMapAPIVersion  = "v1"
)

func CreateKubeConfig(username string) error {

	k8sClient := client.ClientSets().K8s().Kubernetes()

	configName := fmt.Sprintf(kubeconfigNameFormat, username)
	_, err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metav1.GetOptions{})

	if errors.IsNotFound(err) {
		kubeconfig, err := createKubeConfig(username)
		if err != nil {
			klog.Errorln(err)
			return err
		}
		data := map[string]string{fileName: string(kubeconfig)}
		cm := &corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: configMapKind, APIVersion: configMapAPIVersion}, ObjectMeta: metav1.ObjectMeta{Name: configName}, Data: data}
		_, err = k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Create(cm)
		if err != nil && !errors.IsAlreadyExists(err) {
			klog.Errorln(err)
			return err
		}
	}

	return nil
}

func createKubeConfig(username string) ([]byte, error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	kubeconfig := client.ClientSets().K8s().Config()

	ca := kubeconfig.CAData

	csrConfig := &certutil.Config{
		CommonName:   username,
		Organization: nil,
		AltNames:     certutil.AltNames{},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	x509csr, x509key, err := pkiutil.NewCSRAndKey(csrConfig)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	var csrBuffer, keyBuffer bytes.Buffer
	pem.Encode(&keyBuffer, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(x509key)})

	csrBytes, _ := x509.CreateCertificateRequest(rand.Reader, x509csr, x509key)
	pem.Encode(&csrBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	csr := csrBuffer.Bytes()
	key := keyBuffer.Bytes()

	csrName := fmt.Sprintf("%s-csr-%d", username, time.Now().Unix())

	k8sCSR := &certificatesv1beta1.CertificateSigningRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: csrName,
		},
		Spec: certificatesv1beta1.CertificateSigningRequestSpec{
			Request:  csr,
			Usages:   []certificatesv1beta1.KeyUsage{certificatesv1beta1.UsageServerAuth, certificatesv1beta1.UsageKeyEncipherment, certificatesv1beta1.UsageClientAuth, certificatesv1beta1.UsageDigitalSignature},
			Username: username,
			Groups:   []string{"system:authenticated"},
		},
	}

	// create csr
	k8sCSR, err = k8sClient.CertificatesV1beta1().CertificateSigningRequests().Create(k8sCSR)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	// release csr, if it fails need to delete it manually
	defer func() {
		err := k8sClient.CertificatesV1beta1().CertificateSigningRequests().Delete(csrName, &metav1.DeleteOptions{})
		if err != nil {
			klog.Errorln(err)
		}
	}()

	k8sCSR.Status = certificatesv1beta1.CertificateSigningRequestStatus{
		Conditions: []certificatesv1beta1.CertificateSigningRequestCondition{{
			Type:    "Approved",
			Reason:  "KubeSphereApprove",
			Message: "This CSR was approved by KubeSphere certificate approve.",
			LastUpdateTime: metav1.Time{
				Time: time.Now(),
			},
		}},
	}

	// approve csr
	k8sCSR, err = k8sClient.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(k8sCSR)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	// get client cert
	var cert []byte
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {

		k8sCSR, err = k8sClient.CertificatesV1beta1().CertificateSigningRequests().Get(csrName, metav1.GetOptions{})

		if k8sCSR != nil && k8sCSR.Status.Certificate != nil {
			cert = k8sCSR.Status.Certificate
			break
		}

		// sleep 0/200/400 millisecond
		time.Sleep(200 * time.Millisecond * time.Duration(i))
	}

	if cert == nil {
		return nil, fmt.Errorf("create client certificate failed: %v", err)
	}

	currentContext := fmt.Sprintf("%s@%s", username, defaultClusterName)

	config := clientcmdapi.Config{
		Kind:        configMapKind,
		APIVersion:  configMapAPIVersion,
		Preferences: clientcmdapi.Preferences{},
		Clusters: map[string]*clientcmdapi.Cluster{defaultClusterName: {
			Server:                   kubeconfig.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: ca,
		}},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{username: {
			ClientCertificateData: cert,
			ClientKeyData:         key,
		}},
		Contexts: map[string]*clientcmdapi.Context{currentContext: {
			Cluster:   defaultClusterName,
			AuthInfo:  username,
			Namespace: defaultNamespace,
		}},
		CurrentContext: currentContext,
	}

	return clientcmd.Write(config)
}

func GetKubeConfig(username string) (string, error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	configName := fmt.Sprintf(kubeconfigNameFormat, username)
	configMap, err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metav1.GetOptions{})
	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	data := []byte(configMap.Data[fileName])

	kubeconfig, err := clientcmd.Load(data)

	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	masterURL := client.ClientSets().K8s().Master()

	if cluster := kubeconfig.Clusters[defaultClusterName]; cluster != nil {
		cluster.Server = masterURL
	}

	data, err = clientcmd.Write(*kubeconfig)

	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	return string(data), nil
}

func DelKubeConfig(username string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	configName := fmt.Sprintf(kubeconfigNameFormat, username)

	deletePolicy := metav1.DeletePropagationBackground
	err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Delete(configName, &metav1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil {
		klog.Errorln(err)
		return err
	}
	return nil
}

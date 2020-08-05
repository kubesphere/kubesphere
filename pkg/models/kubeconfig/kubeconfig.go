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
	"io/ioutil"
	certificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apiserver/pkg/authentication/user"
	corev1informers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	iamv1alpha2 "kubesphere.io/kubesphere/pkg/apis/iam/v1alpha2"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/scheme"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/utils/pkiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

const (
	inClusterCAFilePath  = "/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	configMapPrefix      = "kubeconfig-"
	kubeconfigNameFormat = configMapPrefix + "%s"
	defaultClusterName   = "local"
	defaultNamespace     = "default"
	kubeconfigFileName   = "config"
	configMapKind        = "ConfigMap"
	configMapAPIVersion  = "v1"
)

type Interface interface {
	GetKubeConfig(username string) (string, error)
	CreateKubeConfig(user *iamv1alpha2.User) error
	UpdateKubeconfig(username string, certificate []byte) error
}

type operator struct {
	k8sClient         kubernetes.Interface
	configMapInformer corev1informers.ConfigMapInformer
	config            *rest.Config
	masterURL         string
}

func NewOperator(k8sClient kubernetes.Interface, configMapInformer corev1informers.ConfigMapInformer, config *rest.Config) Interface {
	return &operator{k8sClient: k8sClient, configMapInformer: configMapInformer, config: config}
}

func NewReadOnlyOperator(configMapInformer corev1informers.ConfigMapInformer, masterURL string) Interface {
	return &operator{configMapInformer: configMapInformer, masterURL: masterURL}
}

func (o *operator) CreateKubeConfig(user *iamv1alpha2.User) error {

	configName := fmt.Sprintf(kubeconfigNameFormat, user.Name)

	_, err := o.configMapInformer.Lister().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName)

	// already exist
	if err == nil {
		return nil
	}

	// internal error
	if !errors.IsNotFound(err) {
		klog.Error(err)
		return err
	}

	// create if not exist
	var ca []byte
	if len(o.config.CAData) > 0 {
		ca = o.config.CAData
	} else {
		ca, err = ioutil.ReadFile(inClusterCAFilePath)
		if err != nil {
			klog.Errorln(err)
			return err
		}
	}

	clientKey, err := o.createCSR(user.Name)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	currentContext := fmt.Sprintf("%s@%s", user.Name, defaultClusterName)

	config := clientcmdapi.Config{
		Kind:        configMapKind,
		APIVersion:  configMapAPIVersion,
		Preferences: clientcmdapi.Preferences{},
		Clusters: map[string]*clientcmdapi.Cluster{defaultClusterName: {
			Server:                   o.config.Host,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: ca,
		}},
		AuthInfos: map[string]*clientcmdapi.AuthInfo{user.Name: {
			ClientKeyData: clientKey,
		}},
		Contexts: map[string]*clientcmdapi.Context{currentContext: {
			Cluster:   defaultClusterName,
			AuthInfo:  user.Name,
			Namespace: defaultNamespace,
		}},
		CurrentContext: currentContext,
	}

	kubeconfig, err := clientcmd.Write(config)

	if err != nil {
		klog.Error(err)
		return err
	}

	cm := &corev1.ConfigMap{TypeMeta: metav1.TypeMeta{Kind: configMapKind, APIVersion: configMapAPIVersion},
		ObjectMeta: metav1.ObjectMeta{Name: configName, Labels: map[string]string{constants.UsernameLabelKey: user.Name}},
		Data:       map[string]string{kubeconfigFileName: string(kubeconfig)}}

	err = controllerutil.SetControllerReference(user, cm, scheme.Scheme)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	_, err = o.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Create(cm)

	if err != nil {
		klog.Errorln(err)
		return err
	}

	return nil
}

func (o *operator) GetKubeConfig(username string) (string, error) {
	configName := fmt.Sprintf(kubeconfigNameFormat, username)
	configMap, err := o.configMapInformer.Lister().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName)
	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	data := []byte(configMap.Data[kubeconfigFileName])

	kubeconfig, err := clientcmd.Load(data)

	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	masterURL := o.masterURL

	// server host override
	if cluster := kubeconfig.Clusters[defaultClusterName]; cluster != nil && masterURL != "" {
		cluster.Server = masterURL
	}

	data, err = clientcmd.Write(*kubeconfig)

	if err != nil {
		klog.Errorln(err)
		return "", err
	}

	return string(data), nil
}

func (o *operator) createCSR(username string) ([]byte, error) {
	csrConfig := &certutil.Config{
		CommonName:   username,
		Organization: nil,
		AltNames:     certutil.AltNames{},
		Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
	x509csr, x509key, err := pkiutil.NewCSRAndKey(csrConfig)
	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	var csrBuffer, keyBuffer bytes.Buffer

	err = pem.Encode(&keyBuffer, &pem.Block{Type: "PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(x509key)})

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, x509csr, x509key)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	err = pem.Encode(&csrBuffer, &pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	csr := csrBuffer.Bytes()
	key := keyBuffer.Bytes()

	csrName := fmt.Sprintf("%s-csr-%d", username, time.Now().Unix())

	k8sCSR := &certificatesv1beta1.CertificateSigningRequest{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CertificateSigningRequest",
			APIVersion: "certificates.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   csrName,
			Labels: map[string]string{constants.UsernameLabelKey: username},
		},
		Spec: certificatesv1beta1.CertificateSigningRequestSpec{
			Request:  csr,
			Usages:   []certificatesv1beta1.KeyUsage{certificatesv1beta1.UsageKeyEncipherment, certificatesv1beta1.UsageClientAuth, certificatesv1beta1.UsageDigitalSignature},
			Username: username,
			Groups:   []string{user.AllAuthenticated},
		},
	}

	// create csr
	k8sCSR, err = o.k8sClient.CertificatesV1beta1().CertificateSigningRequests().Create(k8sCSR)

	if err != nil {
		klog.Errorln(err)
		return nil, err
	}

	return key, nil
}

func (o *operator) UpdateKubeconfig(username string, certificate []byte) error {
	configName := fmt.Sprintf(kubeconfigNameFormat, username)
	configMap, err := o.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metav1.GetOptions{})
	if err != nil {
		klog.Errorln(err)
		return err
	}

	configMap = appendCert(configMap, certificate)
	_, err = o.k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Update(configMap)
	if err != nil {
		klog.Errorln(err)
		return err
	}
	return nil
}

func appendCert(cm *corev1.ConfigMap, cert []byte) *corev1.ConfigMap {
	data := []byte(cm.Data[kubeconfigFileName])

	kubeconfig, err := clientcmd.Load(data)

	// ignore if invalid format
	if err != nil {
		klog.Warning(err)
		return cm
	}

	username := getControlledUsername(cm)

	if kubeconfig.AuthInfos[username] != nil {
		kubeconfig.AuthInfos[username].ClientCertificateData = cert
	}

	data, err = clientcmd.Write(*kubeconfig)

	// ignore if invalid format
	if err != nil {
		klog.Warning(err)
		return cm
	}

	cm.Data[kubeconfigFileName] = string(data)

	return cm
}

func getControlledUsername(cm *corev1.ConfigMap) string {
	for _, ownerReference := range cm.OwnerReferences {
		if ownerReference.Kind == iamv1alpha2.ResourceKindUser {
			return ownerReference.Name
		}
	}
	return ""
}

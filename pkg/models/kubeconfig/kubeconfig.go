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
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/klog"
	"kubesphere.io/kubesphere/pkg/simple/client"
	"math/big"
	rd "math/rand"
	"time"

	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/api/core/v1"

	"kubesphere.io/kubesphere/pkg/constants"
)

const (
	caPath           = "/etc/kubernetes/pki/ca.crt"
	keyPath          = "/etc/kubernetes/pki/ca.key"
	clusterName      = "kubernetes"
	kubectlConfigKey = "config"
	defaultNamespace = "default"
)

type clusterInfo struct {
	CertificateAuthorityData string `yaml:"certificate-authority-data"`
	Server                   string `yaml:"server"`
}

type cluster struct {
	Cluster clusterInfo `yaml:"cluster"`
	Name    string      `yaml:"name"`
}

type contextInfo struct {
	Cluster   string `yaml:"cluster"`
	User      string `yaml:"user"`
	NameSpace string `yaml:"namespace"`
}

type contextObject struct {
	Context contextInfo `yaml:"context"`
	Name    string      `yaml:"name"`
}

type userInfo struct {
	CaData  string `yaml:"client-certificate-data"`
	KeyData string `yaml:"client-key-data"`
}

type user struct {
	Name string   `yaml:"name"`
	User userInfo `yaml:"user"`
}

type kubeConfig struct {
	ApiVersion     string            `yaml:"apiVersion"`
	Clusters       []cluster         `yaml:"clusters"`
	Contexts       []contextObject   `yaml:"contexts"`
	CurrentContext string            `yaml:"current-context"`
	Kind           string            `yaml:"kind"`
	Preferences    map[string]string `yaml:"preferences"`
	Users          []user            `yaml:"users"`
}

type CertInformation struct {
	Country            []string
	Organization       []string
	OrganizationalUnit []string
	EmailAddress       []string
	Province           []string
	Locality           []string
	CommonName         string
	CrtName, KeyName   string
	IsCA               bool
	Names              []pkix.AttributeTypeAndValue
}

func createCRT(RootCa *x509.Certificate, RootKey *rsa.PrivateKey, info CertInformation) ([]byte, []byte, error) {
	var cert, key bytes.Buffer
	Crt := newCertificate(info)
	Key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	var buf []byte

	buf, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)

	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}
	pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: buf})

	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}

	buf = x509.MarshalPKCS1PrivateKey(Key)
	pem.Encode(&key, &pem.Block{Type: "PRIVATE KEY", Bytes: buf})

	return cert.Bytes(), key.Bytes(), nil
}

func Parse(crtPath, keyPath string) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootcertificate, err = parseCrt(crtPath)
	if err != nil {
		klog.Error(err)
		return nil, nil, err
	}
	rootPrivateKey, err = parseKey(keyPath)
	return rootcertificate, rootPrivateKey, nil
}

func parseCrt(path string) (*x509.Certificate, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	p := &pem.Block{}
	p, buf = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

func parseKey(path string) (*rsa.PrivateKey, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		klog.Error(err)
		return nil, err
	}
	p, buf := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func newCertificate(info CertInformation) *x509.Certificate {
	rd.Seed(time.Now().UnixNano())
	return &x509.Certificate{
		SerialNumber: big.NewInt(rd.Int63()),
		Subject: pkix.Name{
			Country:            info.Country,
			Organization:       info.Organization,
			OrganizationalUnit: info.OrganizationalUnit,
			Province:           info.Province,
			CommonName:         info.CommonName,
			Locality:           info.Locality,
			ExtraNames:         info.Names,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  info.IsCA,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses:        info.EmailAddress,
	}
}

func generateCaAndKey(user, caPath, keyPath string) (string, string, error) {
	crtInfo := CertInformation{CommonName: user, IsCA: false}

	crt, pri, err := Parse(caPath, keyPath)
	if err != nil {
		klog.Error(err)
		return "", "", err
	}
	cert, key, err := createCRT(crt, pri, crtInfo)
	if err != nil {
		klog.Error(err)
		return "", "", err
	}

	base64Cert := base64.StdEncoding.EncodeToString(cert)
	base64Key := base64.StdEncoding.EncodeToString(key)
	return base64Cert, base64Key, nil
}

func createKubeConfig(username string) (string, error) {
	tmpKubeConfig := kubeConfig{ApiVersion: "v1", Kind: "Config"}
	serverCa, err := ioutil.ReadFile(caPath)
	if err != nil {
		klog.Errorln(err)
		return "", err
	}
	base64ServerCa := base64.StdEncoding.EncodeToString(serverCa)
	tmpClusterInfo := clusterInfo{CertificateAuthorityData: base64ServerCa, Server: client.ClientSets().K8s().Master()}
	tmpCluster := cluster{Cluster: tmpClusterInfo, Name: clusterName}
	tmpKubeConfig.Clusters = append(tmpKubeConfig.Clusters, tmpCluster)

	contextName := username + "@" + clusterName
	tmpContext := contextObject{Context: contextInfo{User: username, Cluster: clusterName, NameSpace: defaultNamespace}, Name: contextName}
	tmpKubeConfig.Contexts = append(tmpKubeConfig.Contexts, tmpContext)

	cert, key, err := generateCaAndKey(username, caPath, keyPath)

	if err != nil {
		return "", err
	}

	tmpUser := user{User: userInfo{CaData: cert, KeyData: key}, Name: username}
	tmpKubeConfig.Users = append(tmpKubeConfig.Users, tmpUser)
	tmpKubeConfig.CurrentContext = contextName

	config, err := yaml.Marshal(tmpKubeConfig)
	if err != nil {
		return "", err
	}

	return string(config), nil
}

func CreateKubeConfig(username string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	configName := fmt.Sprintf("kubeconfig-%s", username)
	_, err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metaV1.GetOptions{})

	if errors.IsNotFound(err) {
		config, err := createKubeConfig(username)
		if err != nil {
			klog.Errorln(err)
			return err
		}

		data := map[string]string{"config": config}
		configMap := v1.ConfigMap{TypeMeta: metaV1.TypeMeta{Kind: "Configmap", APIVersion: "v1"}, ObjectMeta: metaV1.ObjectMeta{Name: configName}, Data: data}
		_, err = k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Create(&configMap)
		if err != nil && !errors.IsAlreadyExists(err) {
			klog.Errorf("create username %s's kubeConfig failed, reason: %v", username, err)
			return err
		}
	}

	return nil

}

func GetKubeConfig(username string) (string, error) {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	configName := fmt.Sprintf("kubeconfig-%s", username)
	configMap, err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metaV1.GetOptions{})
	if err != nil {
		klog.Errorf("cannot get username %s's kubeConfig, reason: %v", username, err)
		return "", err
	}

	str := configMap.Data[kubectlConfigKey]
	var kubeConfig kubeConfig
	err = yaml.Unmarshal([]byte(str), &kubeConfig)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	masterURL := client.ClientSets().K8s().Master()
	for i, cluster := range kubeConfig.Clusters {
		cluster.Cluster.Server = masterURL
		kubeConfig.Clusters[i] = cluster
	}
	data, err := yaml.Marshal(kubeConfig)
	if err != nil {
		klog.Error(err)
		return "", err
	}
	return string(data), nil
}

func DelKubeConfig(username string) error {
	k8sClient := client.ClientSets().K8s().Kubernetes()
	configName := fmt.Sprintf("kubeconfig-%s", username)
	_, err := k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Get(configName, metaV1.GetOptions{})
	if errors.IsNotFound(err) {
		return nil
	}

	deletePolicy := metaV1.DeletePropagationBackground
	err = k8sClient.CoreV1().ConfigMaps(constants.KubeSphereControlNamespace).Delete(configName, &metaV1.DeleteOptions{PropagationPolicy: &deletePolicy})
	if err != nil {
		klog.Errorf("delete username %s's kubeConfig failed, reason: %v", username, err)
		return err
	}
	return nil
}

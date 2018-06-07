/*
Copyright 2018 The KubeSphere Authors.

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

package models

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/client"
	"kubesphere.io/kubesphere/pkg/options"
	"math/big"
	rd "math/rand"
	"time"
)

const (
	caPath           = "/etc/kubernetes/pki/ca.crt"
	keyPath          = "/etc/kubernetes/pki/ca.key"
	clusterName      = "kubernetes"
	kubectlNamespace = "kubesphere"
	kubectlConfigKey = "config"
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
	Cluster string `yaml:"cluster"`
	User    string `yaml:"user"`
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
		glog.Error(err)
		return nil, nil, err
	}

	var buf []byte

	buf, err = x509.CreateCertificate(rand.Reader, Crt, RootCa, &Key.PublicKey, RootKey)

	if err != nil {
		glog.Error(err)
		return nil, nil, err
	}
	pem.Encode(&cert, &pem.Block{Type: "CERTIFICATE", Bytes: buf})

	if err != nil {
		glog.Error(err)
		return nil, nil, err
	}

	buf = x509.MarshalPKCS1PrivateKey(Key)
	pem.Encode(&key, &pem.Block{Type: "PRIVATE KEY", Bytes: buf})

	return cert.Bytes(), key.Bytes(), nil
}

func Parse(crtPath, keyPath string) (rootcertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootcertificate, err = parseCrt(crtPath)
	if err != nil {
		glog.Error(err)
		return nil, nil, err
	}
	rootPrivateKey, err = parseKey(keyPath)
	return rootcertificate, rootPrivateKey, nil
}

func parseCrt(path string) (*x509.Certificate, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Error(err)
		return nil, err
	}
	p := &pem.Block{}
	p, buf = pem.Decode(buf)
	return x509.ParseCertificate(p.Bytes)
}

func parseKey(path string) (*rsa.PrivateKey, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		glog.Error(err)
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
		IsCA:           info.IsCA,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses: info.EmailAddress,
	}
}

func generateCaAndKey(user, caPath, keyPath string) (string, string, error) {
	crtinfo := CertInformation{CommonName: user, IsCA: false}

	crt, pri, err := Parse(caPath, keyPath)
	if err != nil {
		glog.Error(err)
		return "", "", err
	}
	cert, key, err := createCRT(crt, pri, crtinfo)
	if err != nil {
		glog.Error(err)
		return "", "", err
	}

	base64Cert := base64.StdEncoding.EncodeToString(cert)
	base64Key := base64.StdEncoding.EncodeToString(key)
	return base64Cert, base64Key, nil
}

func createKubeConfig(userName string) (string, error) {
	tmpKubeConfig := kubeConfig{ApiVersion: "v1", Kind: "Config"}
	serverCa, err := ioutil.ReadFile(caPath)
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	base64ServerCa := base64.StdEncoding.EncodeToString(serverCa)
	tmpClusterInfo := clusterInfo{CertificateAuthorityData: base64ServerCa, Server: options.ServerOptions.GetApiServerHost()}
	tmpCluster := cluster{Cluster: tmpClusterInfo, Name: clusterName}
	tmpKubeConfig.Clusters = append(tmpKubeConfig.Clusters, tmpCluster)

	contextName := userName + "@" + clusterName
	tmpContext := contextObject{Context: contextInfo{User: userName, Cluster: clusterName}, Name: contextName}
	tmpKubeConfig.Contexts = append(tmpKubeConfig.Contexts, tmpContext)

	cert, key, err := generateCaAndKey(userName, caPath, keyPath)

	if err != nil {
		return "", err
	}

	tmpUser := user{User: userInfo{CaData: cert, KeyData: key}, Name: userName}
	tmpKubeConfig.Users = append(tmpKubeConfig.Users, tmpUser)
	tmpKubeConfig.CurrentContext = contextName

	config, err := yaml.Marshal(tmpKubeConfig)
	if err != nil {
		return "", err
	}

	return string(config), nil
}

func CreateKubeConfig(user string) error {
	k8sClient := client.NewK8sClient()
	config, err := createKubeConfig(user)
	if err != nil {
		glog.Errorln(err)
		return err
	}

	data := map[string]string{"config": string(config)}
	var configmap = v1.ConfigMap{metav1.TypeMeta{Kind: "Configmap", APIVersion: "v1"}, metav1.ObjectMeta{Name: user}, data}
	_, err = k8sClient.CoreV1().ConfigMaps(kubectlNamespace).Create(&configmap)
	if err != nil {
		glog.Errorln(err)
		return err
	}
	return nil

}

func GetKubeConfig(user string) (string, error) {
	k8sClient := client.NewK8sClient()
	configmap, err := k8sClient.CoreV1().ConfigMaps(kubectlNamespace).Get(user, metav1.GetOptions{})
	if err != nil {
		glog.Errorln(err)
		return "", err
	}
	return configmap.Data[kubectlConfigKey], nil
}

func DelKubeConfig(user string) error {
	k8sClient := client.NewK8sClient()
	err := k8sClient.CoreV1().ConfigMaps(kubectlNamespace).Delete(user, &metav1.DeleteOptions{})
	if err != nil {
		glog.Errorln(err)
		return err
	}
	return nil
}

/*
Copyright 2018 The Kubesphere Authors.

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

// Package options contains flags and options for initializing an apiserver
package options

import (
	goflag "flag"
	"net"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// High enough QPS to fit all expected use cases. QPS=0 is not set here, because
	// client code is overriding it.
	DefaultQPS = 1e6
	// High enough Burst to fit all expected use cases. Burst=0 is not set here, because
	// client code is overriding it.
	DefaultBurst = 1e6
)

// ServerRunOptions runs a kubernetes api server.
type ServerRunOptions struct {
	apiServerHost       string
	insecurePort        int
	port                int
	insecureBindAddress net.IP
	bindAddress         net.IP
	certFile            string
	keyFile             string
	kubeConfigFile      string
	etcdEndpoints       string
	etcdCertFile        string
	etcdKeyFile         string
	etcdCaFile          string
	kubectlImage        string
	mysqlUser           string
	mysqlPasswd         string
	mysqlAddress        string
	opAddress           string
}

// NewServerRunOptions creates a new ServerRunOptions object with default parameters
func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{}
	return &s
}

// AddFlags adds flags for a specific APIServer to the specified FlagSet
func (s *ServerRunOptions) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.apiServerHost, "apiserver-host", "", "The address of the Kubernetes"+
		"Apiserver to connect to in the format of protocol://address:port, e.g., "+
		"http://localhost:8080. If not specified, the assumption is that the binary runs inside")

	fs.IntVar(&s.insecurePort, "insecure-port", 9090, "The port to listen to for incoming HTTP requests.")

	fs.IntVar(&s.port, "port", 8443, "The secure port to listen to for incoming HTTPS requests.")

	fs.IPVar(&s.insecureBindAddress, "insecure-bind-address", net.IPv4(0, 0, 0, 0),
		"The IP address on which to serve the --port (set to 0.0.0.0 for all interfaces).")

	fs.IPVar(&s.bindAddress, "bind-address", net.IPv4(0, 0, 0, 0),
		"The IP address on which to serve the --secure-port (set to 0.0.0.0 for all interfaces).")

	fs.StringVar(&s.certFile, "tls-cert-file", "",
		"File containing the default x509 Certificate for HTTPS.")

	fs.StringVar(&s.keyFile, "tls-key-file", "",
		"File containing the default x509 private key matching --tls-cert-file.")

	fs.StringVar(&s.kubeConfigFile, "kubeconfig", "",
		"Path to kubeconfig file with authorization and master location information.")

	fs.StringVar(&s.etcdEndpoints, "etcd-endpoints", "",
		"Server addresses of etcd")
	fs.StringVar(&s.etcdCertFile, "etcd-tls-cert-file", "",
		"Cert File use to connect etcd in https mode.")

	fs.StringVar(&s.etcdKeyFile, "etcd-tls-key-file", "",
		"Privatekey File use to connect etcd in https mode.")

	fs.StringVar(&s.etcdCaFile, "etcd-tls-ca-file", "",
		"CA Fileuse to connect etcd in https mode.")

	fs.StringVar(&s.kubectlImage, "kubectl-image", "kubectl:1.0",
		"kubectl pod's image")
	fs.StringVar(&s.mysqlAddress, "mysql-addr", "127.0.0.1:3306",
		"Address of mysql, exp:127.0.0.1:3306.")

	fs.StringVar(&s.mysqlPasswd, "mysql-password", "password",
		"Password of mysql")

	fs.StringVar(&s.mysqlUser, "mysql-user", "root",
		"User of mysql.")

	fs.StringVar(&s.opAddress, "openpitrix-address", "http://openpitrix-api-gateway.openpitrix-system.svc",
		"Address of openPitrix")
}

func (s *ServerRunOptions) GetApiServerHost() string {
	return s.apiServerHost
}

func (s *ServerRunOptions) GetInsecurePort() int {
	return s.insecurePort
}

func (s *ServerRunOptions) GetPort() int {
	return s.port
}

func (s *ServerRunOptions) GetInsecureBindAddress() net.IP {
	return s.insecureBindAddress
}

func (s *ServerRunOptions) GetBindAddress() net.IP {
	return s.bindAddress
}

func (s *ServerRunOptions) GetCertFile() string {
	return s.certFile
}

func (s *ServerRunOptions) GetKeyFile() string {
	return s.keyFile
}

func (s *ServerRunOptions) GetKubeConfigFile() string {
	return s.kubeConfigFile
}

func (s *ServerRunOptions) GetKubeConfig() (kubeConfig *rest.Config, err error) {

	kubeConfigFile := s.kubeConfigFile

	if len(kubeConfigFile) > 0 {

		kubeConfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigFile)
		if err != nil {
			return nil, err
		}

	} else {

		kubeConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	kubeConfig.QPS = DefaultQPS
	kubeConfig.Burst = DefaultBurst

	return kubeConfig, nil

}

func (s *ServerRunOptions) GetEtcdEndPoints() []string {
	endpoints := strings.Split(s.etcdEndpoints, ",")
	for k, v := range endpoints {
		endpoints[k] = strings.TrimSpace(v)
	}
	return endpoints
}

func (s *ServerRunOptions) GetEtcdCertFile() string {
	return s.etcdCertFile
}

func (s *ServerRunOptions) GetEtcdKeyFile() string {
	return s.etcdKeyFile
}

func (s *ServerRunOptions) GetEtcdCaFile() string {
	return s.etcdCaFile
}

func (s *ServerRunOptions) GetKubectlImage() string {
	return s.kubectlImage
}

func (s *ServerRunOptions) GetMysqlAddr() string {
	return s.mysqlAddress
}

func (s *ServerRunOptions) GetMysqlUser() string {
	return s.mysqlUser
}

func (s *ServerRunOptions) GetMysqlPassword() string {
	return s.mysqlPasswd
}

func (s *ServerRunOptions) GetOpAddress() string {
	return s.opAddress
}

var ServerOptions = NewServerRunOptions()

func AddFlags(fs *pflag.FlagSet) {
	ServerOptions.addFlags(fs)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

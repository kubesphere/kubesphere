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
	"github.com/spf13/pflag"
	"net"
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

var ServerOptions = NewServerRunOptions()

func AddFlags(fs *pflag.FlagSet) {
	ServerOptions.addFlags(fs)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
}

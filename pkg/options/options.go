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
package options

import (
	"github.com/spf13/pflag"
)

var SharedOptions = NewServerRunOptions()

type ServerRunOptions struct {
	// server bind address
	BindAddress string

	// insecure port number
	InsecurePort int

	// secure port number
	SecurePort int

	// tls cert file
	TlsCertFile string

	// tls private key file
	TlsPrivateKey string

	CommandLine *pflag.FlagSet
}

func NewServerRunOptions() *ServerRunOptions {
	// create default server run options
	s := ServerRunOptions{
		BindAddress:   "0.0.0.0",
		InsecurePort:  9090,
		SecurePort:    0,
		TlsCertFile:   "",
		TlsPrivateKey: "",
		CommandLine:   &pflag.FlagSet{},
	}

	s.CommandLine.StringVar(&s.BindAddress, "bind-address", "0.0.0.0", "server bind address")
	s.CommandLine.IntVar(&s.InsecurePort, "insecure-port", 9090, "insecure port number")
	s.CommandLine.IntVar(&s.SecurePort, "secure-port", 0, "secure port number")
	s.CommandLine.StringVar(&s.TlsCertFile, "tls-cert-file", "", "tls cert file")
	s.CommandLine.StringVar(&s.TlsPrivateKey, "tls-private-key", "", "tls private key")

	return &s
}

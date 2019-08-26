// Copyright (c) 2019 Tigera, Inc. All rights reserved.

// This code has been based on code from etcd repository
// to provide support for inline certificates and keys for calicoctl.
// Below are the github links for the files from which the code has been borrowed.

// Copyright 2015 The etcd Authors
//     https://github.com/etcd-io/etcd/blob/release-3.3/pkg/transport/listener.go

// Copyright 2016 The etcd Authors
//     https://github.com/etcd-io/etcd/blob/release-3.3/pkg/tlsutil/tlsutil.go

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package etcdv3

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// This struct is used to store the inline cert, key and CA cert data
type TlsInlineCertKey struct {
	Cert   string
	Key    string
	CACert string
}

// ClientConfigInlineCertKey() returns a pointer to tls.Config struct object with certificate data
// for client creation using only the inline certificate, key and CA certificate data.
func (info TlsInlineCertKey) ClientConfigInlineCertKey() (*tls.Config, error) {
	var cfg *tls.Config
	var err error

	if info.Cert == "" || info.Key == "" {
		return nil, fmt.Errorf("Certificate and Key must both be present inline.")
	}

	if info.Cert != "" && info.Key != "" {
		cfg, err = info.baseCertConfig()
		if err != nil {
			return nil, err
		}
	}

	if info.CACert != "" {
		cfg.RootCAs, err = newCertPool(info.CACert)
		if err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

// baseCertConfig() populates tls struct with certificate data
func (info TlsInlineCertKey) baseCertConfig() (*tls.Config, error) {
	_, err := newCert([]byte(info.Cert), []byte(info.Key))
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	cfg.GetCertificate = func(clientHello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		return newCert([]byte(info.Cert), []byte(info.Key))
	}
	cfg.GetClientCertificate = func(unused *tls.CertificateRequestInfo) (*tls.Certificate, error) {
		return newCert([]byte(info.Cert), []byte(info.Key))
	}
	return cfg, nil
}

// newCertPool() creates the certificate pool from the CA certificates provided
func newCertPool(caCert string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()
	if caCert == "" {
		return nil, nil
	}
	var block *pem.Block
	certByte := []byte(caCert)
	block, certByte = pem.Decode(certByte)
	if block == nil {
		return nil, fmt.Errorf("Cannot decode PEM block containing certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	certPool.AddCert(cert)
	return certPool, nil
}

// newCert() generates TLS cert by using the given cert and key values.
func newCert(cert, key []byte) (*tls.Certificate, error) {
	tlsCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

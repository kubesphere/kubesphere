/*
Copyright 2016 The Kubernetes Authors.

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

package pkiutil

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"io/ioutil"
	"net"
	"os"
	"testing"

	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

func TestNewCertificateAuthority(t *testing.T) {
	cert, key, err := NewCertificateAuthority()

	if cert == nil {
		t.Errorf(
			"failed NewCertificateAuthority, cert == nil",
		)
	}
	if key == nil {
		t.Errorf(
			"failed NewCertificateAuthority, key == nil",
		)
	}
	if err != nil {
		t.Errorf(
			"failed NewCertificateAuthority with an error: %v",
			err,
		)
	}
}

func TestNewCertAndKey(t *testing.T) {
	var tests = []struct {
		caKeySize int
		expected  bool
	}{
		{
			// RSA key too small
			caKeySize: 128,
			expected:  false,
		},
		{
			// Should succeed
			caKeySize: 2048,
			expected:  true,
		},
	}

	for _, rt := range tests {
		caKey, err := rsa.GenerateKey(rand.Reader, rt.caKeySize)
		if err != nil {
			t.Fatalf("Couldn't create rsa Private Key")
		}
		caCert := &x509.Certificate{}
		config := certutil.Config{
			CommonName:   "test",
			Organization: []string{"test"},
			Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		}
		_, _, actual := NewCertAndKey(caCert, caKey, config)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed NewCertAndKey:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestHasServerAuth(t *testing.T) {
	caCert, caKey, _ := NewCertificateAuthority()

	var tests = []struct {
		config   certutil.Config
		expected bool
	}{
		{
			config: certutil.Config{
				CommonName: "test",
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
			expected: true,
		},
		{
			config: certutil.Config{
				CommonName: "test",
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
			expected: false,
		},
	}

	for _, rt := range tests {
		cert, _, err := NewCertAndKey(caCert, caKey, rt.config)
		if err != nil {
			t.Fatalf("Couldn't create cert: %v", err)
		}
		actual := HasServerAuth(cert)
		if actual != rt.expected {
			t.Errorf(
				"failed HasServerAuth:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				actual,
			)
		}
	}
}

func TestWriteCertAndKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	caCert := &x509.Certificate{}
	actual := WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWriteCert(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert := &x509.Certificate{}
	actual := WriteCert(tmpdir, "foo", caCert)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWriteKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	actual := WriteKey(tmpdir, "foo", caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestWritePublicKey(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	actual := WritePublicKey(tmpdir, "foo", &caKey.PublicKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}
}

func TestCertOrKeyExist(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Couldn't create rsa Private Key")
	}
	caCert := &x509.Certificate{}
	actual := WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if actual != nil {
		t.Errorf(
			"failed WriteCertAndKey with an error: %v",
			actual,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		actual := CertOrKeyExist(rt.path, rt.name)
		if actual != rt.expected {
			t.Errorf(
				"failed CertOrKeyExist:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				actual,
			)
		}
	}
}

func TestTryLoadCertAndKeyFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert, caKey, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteCertAndKey(tmpdir, "foo", caCert, caKey)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, _, actual := TryLoadCertAndKeyFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestTryLoadCertFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	caCert, _, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteCert(tmpdir, "foo", caCert)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, actual := TryLoadCertFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestTryLoadKeyFromDisk(t *testing.T) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("Couldn't create tmpdir")
	}
	defer os.RemoveAll(tmpdir)

	_, caKey, err := NewCertificateAuthority()
	if err != nil {
		t.Errorf(
			"failed to create cert and key with an error: %v",
			err,
		)
	}
	err = WriteKey(tmpdir, "foo", caKey)
	if err != nil {
		t.Errorf(
			"failed to write cert and key with an error: %v",
			err,
		)
	}

	var tests = []struct {
		path     string
		name     string
		expected bool
	}{
		{
			path:     "",
			name:     "",
			expected: false,
		},
		{
			path:     tmpdir,
			name:     "foo",
			expected: true,
		},
	}
	for _, rt := range tests {
		_, actual := TryLoadKeyFromDisk(rt.path, rt.name)
		if (actual == nil) != rt.expected {
			t.Errorf(
				"failed TryLoadCertAndKeyFromDisk:\n\texpected: %t\n\t  actual: %t",
				rt.expected,
				(actual == nil),
			)
		}
	}
}

func TestPathsForCertAndKey(t *testing.T) {
	crtPath, keyPath := pathsForCertAndKey("/foo", "bar")
	if crtPath != "/foo/bar.crt" {
		t.Errorf("unexpected certificate path: %s", crtPath)
	}
	if keyPath != "/foo/bar.key" {
		t.Errorf("unexpected key path: %s", keyPath)
	}
}

func TestPathForCert(t *testing.T) {
	crtPath := pathForCert("/foo", "bar")
	if crtPath != "/foo/bar.crt" {
		t.Errorf("unexpected certificate path: %s", crtPath)
	}
}

func TestPathForKey(t *testing.T) {
	keyPath := pathForKey("/foo", "bar")
	if keyPath != "/foo/bar.key" {
		t.Errorf("unexpected certificate path: %s", keyPath)
	}
}

func TestPathForPublicKey(t *testing.T) {
	pubPath := pathForPublicKey("/foo", "bar")
	if pubPath != "/foo/bar.pub" {
		t.Errorf("unexpected certificate path: %s", pubPath)
	}
}

func TestGetAPIServerAltNames(t *testing.T) {

	var tests = []struct {
		name                string
		cfg                 *kubeadmapi.MasterConfiguration
		expectedDNSNames    []string
		expectedIPAddresses []string
	}{
		{
			name: "ControlPlaneEndpoint DNS",
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{AdvertiseAddress: "1.2.3.4", ControlPlaneEndpoint: "api.k8s.io:6443"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "10.96.0.0/12", DNSDomain: "cluster.local"},
				NodeName:          "valid-hostname",
				APIServerCertSANs: []string{"10.1.245.94", "10.1.245.95", "1.2.3.L", "invalid,commas,in,DNS"},
			},
			expectedDNSNames:    []string{"valid-hostname", "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster.local", "api.k8s.io"},
			expectedIPAddresses: []string{"10.96.0.1", "1.2.3.4", "10.1.245.94", "10.1.245.95"},
		},
		{
			name: "ControlPlaneEndpoint IP",
			cfg: &kubeadmapi.MasterConfiguration{
				API:               kubeadmapi.API{AdvertiseAddress: "1.2.3.4", ControlPlaneEndpoint: "4.5.6.7:6443"},
				Networking:        kubeadmapi.Networking{ServiceSubnet: "10.96.0.0/12", DNSDomain: "cluster.local"},
				NodeName:          "valid-hostname",
				APIServerCertSANs: []string{"10.1.245.94", "10.1.245.95", "1.2.3.L", "invalid,commas,in,DNS"},
			},
			expectedDNSNames:    []string{"valid-hostname", "kubernetes", "kubernetes.default", "kubernetes.default.svc", "kubernetes.default.svc.cluster.local"},
			expectedIPAddresses: []string{"10.96.0.1", "1.2.3.4", "10.1.245.94", "10.1.245.95", "4.5.6.7"},
		},
	}

	for _, rt := range tests {
		altNames, err := GetAPIServerAltNames(rt.cfg)
		if err != nil {
			t.Fatalf("failed calling GetAPIServerAltNames: %s: %v", rt.name, err)
		}

		for _, DNSName := range rt.expectedDNSNames {
			found := false
			for _, val := range altNames.DNSNames {
				if val == DNSName {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("%s: altNames does not contain DNSName %s but %v", rt.name, DNSName, altNames.DNSNames)
			}
		}

		for _, IPAddress := range rt.expectedIPAddresses {
			found := false
			for _, val := range altNames.IPs {
				if val.Equal(net.ParseIP(IPAddress)) {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("%s: altNames does not contain IPAddress %s but %v", rt.name, IPAddress, altNames.IPs)
			}
		}
	}
}

func TestGetEtcdAltNames(t *testing.T) {
	proxy := "user-etcd-proxy"
	proxyIP := "10.10.10.100"
	cfg := &kubeadmapi.MasterConfiguration{
		Etcd: kubeadmapi.Etcd{
			ServerCertSANs: []string{
				proxy,
				proxyIP,
				"1.2.3.L",
				"invalid,commas,in,DNS",
			},
		},
	}

	altNames, err := GetEtcdAltNames(cfg)
	if err != nil {
		t.Fatalf("failed calling GetEtcdAltNames: %v", err)
	}

	expectedDNSNames := []string{"localhost", proxy}
	for _, DNSName := range expectedDNSNames {
		found := false
		for _, val := range altNames.DNSNames {
			if val == DNSName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain DNSName %s", DNSName)
		}
	}

	expectedIPAddresses := []string{"127.0.0.1", proxyIP}
	for _, IPAddress := range expectedIPAddresses {
		found := false
		for _, val := range altNames.IPs {
			if val.Equal(net.ParseIP(IPAddress)) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain IPAddress %s", IPAddress)
		}
	}
}

func TestGetEtcdPeerAltNames(t *testing.T) {
	hostname := "valid-hostname"
	proxy := "user-etcd-proxy"
	proxyIP := "10.10.10.100"
	advertiseIP := "1.2.3.4"
	cfg := &kubeadmapi.MasterConfiguration{
		API:      kubeadmapi.API{AdvertiseAddress: advertiseIP},
		NodeName: hostname,
		Etcd: kubeadmapi.Etcd{
			PeerCertSANs: []string{
				proxy,
				proxyIP,
				"1.2.3.L",
				"invalid,commas,in,DNS",
			},
		},
	}

	altNames, err := GetEtcdPeerAltNames(cfg)
	if err != nil {
		t.Fatalf("failed calling GetEtcdPeerAltNames: %v", err)
	}

	expectedDNSNames := []string{hostname, proxy}
	for _, DNSName := range expectedDNSNames {
		found := false
		for _, val := range altNames.DNSNames {
			if val == DNSName {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain DNSName %s", DNSName)
		}
	}

	expectedIPAddresses := []string{advertiseIP, proxyIP}
	for _, IPAddress := range expectedIPAddresses {
		found := false
		for _, val := range altNames.IPs {
			if val.Equal(net.ParseIP(IPAddress)) {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("altNames does not contain IPAddress %s", IPAddress)
		}
	}
}

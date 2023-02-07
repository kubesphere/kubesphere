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

package util

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "k8s.io/api/core/v1"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"
	"k8s.io/klog/v2"

	fedv1b1 "sigs.k8s.io/kubefed/pkg/apis/core/v1beta1"
	"sigs.k8s.io/kubefed/pkg/client/generic"
)

const (
	DefaultKubeFedSystemNamespace = "kube-federation-system"

	KubeAPIQPS        = 20.0
	KubeAPIBurst      = 30
	TokenKey          = "token"
	CaCrtKey          = "ca.crt"
	KubeFedConfigName = "kubefed"
)

// BuildClusterConfig returns a restclient.Config that can be used to configure
// a client for the given KubeFedCluster or an error. The client is used to
// access kubernetes secrets in the kubefed namespace.
func BuildClusterConfig(fedCluster *fedv1b1.KubeFedCluster, client generic.Client, fedNamespace string) (*restclient.Config, error) {
	clusterName := fedCluster.Name

	apiEndpoint := fedCluster.Spec.APIEndpoint
	// TODO(marun) Remove when validation ensures a non-empty value.
	if apiEndpoint == "" {
		return nil, errors.Errorf("The api endpoint of cluster %s is empty", clusterName)
	}

	secretName := fedCluster.Spec.SecretRef.Name
	if secretName == "" {
		return nil, errors.Errorf("Cluster %s does not have a secret name", clusterName)
	}
	secret := &apiv1.Secret{}
	err := client.Get(context.TODO(), secret, fedNamespace, secretName)
	if err != nil {
		return nil, err
	}

	token, tokenFound := secret.Data[TokenKey]
	if !tokenFound || len(token) == 0 {
		return nil, errors.Errorf("The secret for cluster %s is missing a non-empty value for %q", clusterName, TokenKey)
	}

	clusterConfig, err := clientcmd.BuildConfigFromFlags(apiEndpoint, "")
	if err != nil {
		return nil, err
	}
	clusterConfig.CAData = fedCluster.Spec.CABundle
	clusterConfig.BearerToken = string(token)
	clusterConfig.QPS = KubeAPIQPS
	clusterConfig.Burst = KubeAPIBurst

	if fedCluster.Spec.ProxyURL != "" {
		proxyURL, err := url.Parse(fedCluster.Spec.ProxyURL)
		if err != nil {
			return nil, errors.Errorf("Failed to parse provided proxy URL %s: %v", fedCluster.Spec.ProxyURL, err)
		}
		clusterConfig.Proxy = http.ProxyURL(proxyURL)
	}

	if len(fedCluster.Spec.DisabledTLSValidations) != 0 {
		klog.V(1).Infof("Cluster %s will use a custom transport for TLS certificate validation", fedCluster.Name)
		if err = CustomizeTLSTransport(fedCluster, clusterConfig); err != nil {
			return nil, err
		}
	}

	return clusterConfig, nil
}

// IsPrimaryCluster checks if the caller is working with objects for the
// primary cluster by checking if the UIDs match for both ObjectMetas passed
// in.
// TODO (font): Need to revisit this when cluster ID is available.
func IsPrimaryCluster(obj, clusterObj runtimeclient.Object) bool {
	meta := MetaAccessor(obj)
	clusterMeta := MetaAccessor(clusterObj)
	return meta.GetUID() == clusterMeta.GetUID()
}

// CustomizeTLSTransport replaces the restclient.Config.Transport with one that
// implements the desired TLS certificate validations
func CustomizeTLSTransport(fedCluster *fedv1b1.KubeFedCluster, clientConfig *restclient.Config) error {
	clientTransportConfig, err := clientConfig.TransportConfig()
	if err != nil {
		return errors.Errorf("Cluster %s client transport config error: %s", fedCluster.Name, err)
	}
	transportConfig, err := transport.TLSConfigFor(clientTransportConfig)
	if err != nil {
		return errors.Errorf("Cluster %s transport error: %s", fedCluster.Name, err)
	}

	if transportConfig != nil {
		err = CustomizeCertificateValidation(fedCluster, transportConfig)
		if err != nil {
			return errors.Errorf("Cluster %s custom certificate validation error: %s", fedCluster.Name, err)
		}

		// using the same defaults as http.DefaultTransport
		clientConfig.Transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       transportConfig,
		}
		clientConfig.TLSClientConfig = restclient.TLSClientConfig{}
	} else {
		clientConfig.Insecure = true
	}

	return nil
}

// CustomizeCertificateValidation modifies an existing tls.Config to disable the
// desired TLS checks in KubeFedCluster config
func CustomizeCertificateValidation(fedCluster *fedv1b1.KubeFedCluster, tlsConfig *tls.Config) error {
	// InsecureSkipVerify must be enabled to prevent early validation errors from
	// returning before VerifyPeerCertificate is run
	tlsConfig.InsecureSkipVerify = true

	var ignoreSubjectName, ignoreValidityPeriod bool
	for _, validation := range fedCluster.Spec.DisabledTLSValidations {
		switch fedv1b1.TLSValidation(validation) {
		case fedv1b1.TLSAll:
			klog.V(1).Infof("Cluster %s will not perform TLS certificate validation", fedCluster.Name)
			return nil
		case fedv1b1.TLSSubjectName:
			ignoreSubjectName = true
		case fedv1b1.TLSValidityPeriod:
			ignoreValidityPeriod = true
		}
	}

	// Normal TLS SubjectName validation uses the conn dnsname for validation,
	// but this is not available when using a VerifyPeerCertificate functions.
	// As a workaround, we will fill the tls.Config.ServerName with the URL host
	// specified as the KubeFedCluster API target
	if !ignoreSubjectName && tlsConfig.ServerName == "" {
		apiURL, err := url.Parse(fedCluster.Spec.APIEndpoint)
		if err != nil {
			return errors.Errorf("failed to identify a valid host from APIEndpoint for use in SubjectName validation")
		}
		tlsConfig.ServerName = apiURL.Hostname()
	}

	// VerifyPeerCertificate uses the same logic as crypto/tls Conn.verifyServerCertificate
	// but uses a modified set of options to ignore specific validations
	tlsConfig.VerifyPeerCertificate = func(certificates [][]byte, verifiedChains [][]*x509.Certificate) error {
		opts := x509.VerifyOptions{
			Roots:         tlsConfig.RootCAs,
			CurrentTime:   time.Now(),
			Intermediates: x509.NewCertPool(),
			DNSName:       tlsConfig.ServerName,
		}
		if tlsConfig.Time != nil {
			opts.CurrentTime = tlsConfig.Time()
		}

		certs := make([]*x509.Certificate, len(certificates))
		for i, asn1Data := range certificates {
			cert, err := x509.ParseCertificate(asn1Data)
			if err != nil {
				return errors.New("tls: failed to parse certificate from server: " + err.Error())
			}
			certs[i] = cert
		}

		for i, cert := range certs {
			if i == 0 {
				continue
			}
			opts.Intermediates.AddCert(cert)
		}

		if ignoreSubjectName {
			// set the DNSName to nil to ignore the name validation
			opts.DNSName = ""
			klog.V(1).Infof("Cluster %s will not perform tls certificate SubjectName validation", fedCluster.Name)
		}
		if ignoreValidityPeriod {
			// set the CurrentTime to immediately after the certificate start time
			// this will ensure that certificate passes the validity period check
			opts.CurrentTime = certs[0].NotBefore.Add(time.Second)
			klog.V(1).Infof("Cluster %s will not perform tls certificate ValidityPeriod validation", fedCluster.Name)
		}

		_, err := certs[0].Verify(opts)

		return err
	}

	return nil
}

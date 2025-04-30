/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package helm

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
)

func NewTLSConfig(caBundle string, insecureSkipTLSVerify bool) (*tls.Config, error) {
	config := tls.Config{
		InsecureSkipVerify: insecureSkipTLSVerify,
	}

	if caBundle != "" {
		caCerts, err := base64.StdEncoding.DecodeString(caBundle)
		if err != nil {
			return nil, fmt.Errorf("failed to decode caBundle: %v", err)
		}
		cp, err := certPoolFromCABundle(caCerts)
		if err != nil {
			return nil, err
		}
		config.RootCAs = cp
	}

	return &config, nil
}

func certPoolFromCABundle(caCerts []byte) (*x509.CertPool, error) {
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(caCerts) {
		return nil, errors.Errorf("failed to append certificates from caBundle")
	}
	return cp, nil
}

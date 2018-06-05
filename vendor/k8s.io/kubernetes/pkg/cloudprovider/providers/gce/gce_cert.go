/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	"context"

	compute "google.golang.org/api/compute/v1"

	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/filter"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/gce/cloud/meta"
)

func newCertMetricContext(request string) *metricContext {
	return newGenericMetricContext("cert", request, unusedMetricLabel, unusedMetricLabel, computeV1Version)
}

// GetSslCertificate returns the SslCertificate by name.
func (gce *GCECloud) GetSslCertificate(name string) (*compute.SslCertificate, error) {
	mc := newCertMetricContext("get")
	v, err := gce.c.SslCertificates().Get(context.Background(), meta.GlobalKey(name))
	return v, mc.Observe(err)
}

// CreateSslCertificate creates and returns a SslCertificate.
func (gce *GCECloud) CreateSslCertificate(sslCerts *compute.SslCertificate) (*compute.SslCertificate, error) {
	mc := newCertMetricContext("create")
	err := gce.c.SslCertificates().Insert(context.Background(), meta.GlobalKey(sslCerts.Name), sslCerts)
	if err != nil {
		return nil, mc.Observe(err)
	}
	return gce.GetSslCertificate(sslCerts.Name)
}

// DeleteSslCertificate deletes the SslCertificate by name.
func (gce *GCECloud) DeleteSslCertificate(name string) error {
	mc := newCertMetricContext("delete")
	return mc.Observe(gce.c.SslCertificates().Delete(context.Background(), meta.GlobalKey(name)))
}

// ListSslCertificates lists all SslCertificates in the project.
func (gce *GCECloud) ListSslCertificates() ([]*compute.SslCertificate, error) {
	mc := newCertMetricContext("list")
	v, err := gce.c.SslCertificates().List(context.Background(), filter.None)
	return v, mc.Observe(err)
}

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

func newUrlMapMetricContext(request string) *metricContext {
	return newGenericMetricContext("urlmap", request, unusedMetricLabel, unusedMetricLabel, computeV1Version)
}

// GetUrlMap returns the UrlMap by name.
func (gce *GCECloud) GetUrlMap(name string) (*compute.UrlMap, error) {
	mc := newUrlMapMetricContext("get")
	v, err := gce.c.UrlMaps().Get(context.Background(), meta.GlobalKey(name))
	return v, mc.Observe(err)
}

// CreateUrlMap creates a url map
func (gce *GCECloud) CreateUrlMap(urlMap *compute.UrlMap) error {
	mc := newUrlMapMetricContext("create")
	return mc.Observe(gce.c.UrlMaps().Insert(context.Background(), meta.GlobalKey(urlMap.Name), urlMap))
}

// UpdateUrlMap applies the given UrlMap as an update
func (gce *GCECloud) UpdateUrlMap(urlMap *compute.UrlMap) error {
	mc := newUrlMapMetricContext("update")
	return mc.Observe(gce.c.UrlMaps().Update(context.Background(), meta.GlobalKey(urlMap.Name), urlMap))
}

// DeleteUrlMap deletes a url map by name.
func (gce *GCECloud) DeleteUrlMap(name string) error {
	mc := newUrlMapMetricContext("delete")
	return mc.Observe(gce.c.UrlMaps().Delete(context.Background(), meta.GlobalKey(name)))
}

// ListUrlMaps lists all UrlMaps in the project.
func (gce *GCECloud) ListUrlMaps() ([]*compute.UrlMap, error) {
	mc := newUrlMapMetricContext("list")
	v, err := gce.c.UrlMaps().List(context.Background(), filter.None)
	return v, mc.Observe(err)
}

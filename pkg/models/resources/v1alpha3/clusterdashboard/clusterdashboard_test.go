/*
Copyright 2021 The KubeSphere Authors.

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

package clusterdashboard

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	monitoringdashboardv1alpha1 "kubesphere.io/monitoring-dashboard/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var c client.Client

func compare(actual *monitoringdashboardv1alpha1.ClusterDashboard,
	expects ...*monitoringdashboardv1alpha1.ClusterDashboard) bool {
	for _, app := range expects {
		if actual.Name == app.Name && reflect.DeepEqual(actual.Labels, app.Labels) {
			return true
		}
	}
	return false
}

func TestGetListClusterDashboards(t *testing.T) {
	e := &envtest.Environment{CRDDirectoryPaths: []string{filepath.Join("..", "..", "..", "..", "..", "config", "crds")}}
	cfg, err := e.Start()
	if err != nil {
		t.Fatal(err)
	}

	sch := scheme.Scheme
	if err := monitoringdashboardv1alpha1.AddToScheme(sch); err != nil {
		t.Fatalf("unable add APIs to scheme: %v", err)
	}

	stopCh := make(chan struct{})

	ce, _ := cache.New(cfg, cache.Options{Scheme: sch})
	go ce.Start(stopCh)
	ce.WaitForCacheSync(stopCh)

	c, _ = client.New(cfg, client.Options{Scheme: sch})

	var labelSet1 = map[string]string{"foo-1": "bar-1"}
	var labelSet2 = map[string]string{"foo-2": "bar-2"}

	testCases := []*monitoringdashboardv1alpha1.ClusterDashboard{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clusterdashboard-1",
				Labels: labelSet1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "clusterdashboard-2",
				Labels: labelSet2,
			},
		},
	}

	ctx := context.TODO()

	for _, board := range testCases {
		if err = c.Create(ctx, board); err != nil {
			t.Fatal(err)
		}
	}

	getter := New(ce)

	results, err := getter.List("", &query.Query{})
	if err != nil {
		t.Fatal(err)
	}

	if results.TotalItems != len(testCases) {
		t.Fatal("TotalItems is not match")
	}

	if len(results.Items) != len(testCases) {
		t.Fatal("Items numbers is not match mock data")
	}

	for _, dashboard := range results.Items {
		dashboard, err := dashboard.(*monitoringdashboardv1alpha1.ClusterDashboard)
		if !err {
			t.Fatal(err)
		}
		if !compare(dashboard, testCases...) {
			t.Errorf("The results %v not match testcases %v", results.Items, testCases)
		}
	}

	result, err := getter.Get("", "clusterdashboard-1")
	if err != nil {
		t.Fatal(err)
	}

	dashboard := result.(*monitoringdashboardv1alpha1.ClusterDashboard)
	if !compare(dashboard, testCases...) {
		t.Errorf("The results %v not match testcases %v", result, testCases)
	}
}

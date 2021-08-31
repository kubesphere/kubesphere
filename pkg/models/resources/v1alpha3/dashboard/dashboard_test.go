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

package dashboard

import (
	"context"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	monitoringdashboardv1alpha2 "kubesphere.io/monitoring-dashboard/api/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/apiserver/query"
)

var c client.Client

func compare(actual *monitoringdashboardv1alpha2.Dashboard, expects ...*monitoringdashboardv1alpha2.Dashboard) bool {
	for _, app := range expects {
		if actual.Name == app.Name && actual.Namespace == app.Namespace && reflect.DeepEqual(actual.Labels, app.Labels) {
			return true
		}
	}
	return false
}

func TestGetListDashboards(t *testing.T) {
	sch := scheme.Scheme
	if err := monitoringdashboardv1alpha2.AddToScheme(sch); err != nil {
		t.Fatalf("unable add APIs to scheme: %v", err)
	}

	c = fake.NewFakeClientWithScheme(sch)

	var labelSet1 = map[string]string{"foo-1": "bar-1"}
	var labelSet2 = map[string]string{"foo-2": "bar-2"}

	var ns = "ns-1"
	testCases := []*monitoringdashboardv1alpha2.Dashboard{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dashboard-1",
				Namespace: ns,
				Labels:    labelSet1,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "dashboard-2",
				Namespace: ns,
				Labels:    labelSet2,
			},
		},
	}

	ctx := context.TODO()

	for _, board := range testCases {
		if err := c.Create(ctx, board); err != nil {
			t.Fatal(err)
		}
	}

	getter := New(c)

	results, err := getter.List(ns, &query.Query{})
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
		dashboard, err := dashboard.(*monitoringdashboardv1alpha2.Dashboard)
		if !err {
			t.Fatal(err)
		}
		if !compare(dashboard, testCases...) {
			t.Errorf("The results %v not match testcases %v", results.Items, testCases)
		}
	}

	result, err := getter.Get(ns, "dashboard-1")
	if err != nil {
		t.Fatal(err)
	}

	dashboard := result.(*monitoringdashboardv1alpha2.Dashboard)
	if !compare(dashboard, testCases...) {
		t.Errorf("The results %v not match testcases %v", result, testCases)
	}
}

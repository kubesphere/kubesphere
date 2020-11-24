/*
Copyright 2020 KubeSphere Authors

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

package v1alpha3

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/informers"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"testing"
	"time"
)

func TestIsRangeQuery(t *testing.T) {
	tests := []struct {
		opt      queryOptions
		expected bool
	}{
		{
			opt: queryOptions{
				time: time.Now(),
			},
			expected: false,
		},
		{
			opt: queryOptions{
				start: time.Now().Add(-time.Hour),
				end:   time.Now(),
			},
			expected: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			b := tt.opt.isRangeQuery()
			if b != tt.expected {
				t.Fatalf("expected %v, but got %v", tt.expected, b)
			}
		})
	}
}

func TestParseRequestParams(t *testing.T) {
	tests := []struct {
		params      reqParams
		lvl         monitoring.Level
		namespace   corev1.Namespace
		expected    queryOptions
		expectedErr bool
	}{
		{
			params: reqParams{
				time: "abcdef",
			},
			lvl:         monitoring.LevelCluster,
			expectedErr: true,
		},
		{
			params: reqParams{
				time: "1585831995",
			},
			lvl: monitoring.LevelCluster,
			expected: queryOptions{
				time:         time.Unix(1585831995, 0),
				metricFilter: ".*",
				namedMetrics: model.ClusterMetrics,
				option:       monitoring.ClusterOption{},
			},
			expectedErr: false,
		},
		{
			params: reqParams{
				start:         "1585830000",
				end:           "1585839999",
				step:          "1m",
				namespaceName: "default",
			},
			lvl: monitoring.LevelNamespace,
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Unix(1585836666, 0),
					},
				},
			},
			expected: queryOptions{
				start:        time.Unix(1585836699, 0),
				end:          time.Unix(1585839999, 0),
				step:         time.Minute,
				identifier:   model.IdentifierNamespace,
				metricFilter: ".*",
				namedMetrics: model.NamespaceMetrics,
				option: monitoring.NamespaceOption{
					ResourceFilter: ".*",
					NamespaceName:  "default",
				},
			},
			expectedErr: false,
		},
		{
			params: reqParams{
				time:          "1585830000",
				namespaceName: "default",
			},
			lvl: monitoring.LevelNamespace,
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Unix(1585836666, 0),
					},
				},
			},
			expectedErr: true,
		},
		{
			params: reqParams{
				start:         "1585830000",
				end:           "1585839999",
				step:          "1m",
				namespaceName: "default",
			},
			lvl: monitoring.LevelNamespace,
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Unix(1589999999, 0),
					},
				},
			},
			expectedErr: true,
		},
		{
			params: reqParams{
				start:         "1585830000",
				end:           "1585839999",
				step:          "1m",
				namespaceName: "non-exist",
			},
			lvl: monitoring.LevelNamespace,
			namespace: corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
					CreationTimestamp: metav1.Time{
						Time: time.Unix(1589999999, 0),
					},
				},
			},
			expectedErr: true,
		},
		{
			params: reqParams{
				time:          "1585830000",
				componentType: "etcd",
				metricFilter:  "etcd_server_list",
			},
			lvl: monitoring.LevelComponent,
			expected: queryOptions{
				time:         time.Unix(1585830000, 0),
				metricFilter: "etcd_server_list",
				namedMetrics: model.EtcdMetrics,
				option:       monitoring.ComponentOption{},
			},
			expectedErr: false,
		},
		{
			params: reqParams{
				time:          "1585830000",
				workspaceName: "system-workspace",
				metricFilter:  "namespace_memory_usage_wo_cache|namespace_memory_limit_hard|namespace_cpu_usage",
				page:          "1",
				limit:         "10",
				order:         "desc",
				target:        "namespace_cpu_usage",
			},
			lvl: monitoring.LevelNamespace,
			expected: queryOptions{
				time:         time.Unix(1585830000, 0),
				metricFilter: "namespace_memory_usage_wo_cache|namespace_memory_limit_hard|namespace_cpu_usage",
				namedMetrics: model.NamespaceMetrics,
				option: monitoring.NamespaceOption{
					ResourceFilter: ".*",
					WorkspaceName:  "system-workspace",
				},
				target:     "namespace_cpu_usage",
				identifier: "namespace",
				order:      "desc",
				page:       1,
				limit:      10,
			},
			expectedErr: false,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			client := fake.NewSimpleClientset(&tt.namespace)
			fakeInformerFactory := informers.NewInformerFactories(client, nil, nil, nil, nil, nil)
			handler := newHandler(client, nil, fakeInformerFactory, nil)

			result, err := handler.makeQueryOptions(tt.params, tt.lvl)
			if err != nil {
				if !tt.expectedErr {
					t.Fatalf("unexpected err: %s.", err.Error())
				}
				return
			}

			if tt.expectedErr {
				t.Fatalf("failed to catch error.")
			}

			if diff := cmp.Diff(result, tt.expected, cmp.AllowUnexported(result, tt.expected)); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", tt.expected, diff)
			}
		})
	}
}

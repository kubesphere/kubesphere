package v1alpha3

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	model "kubesphere.io/kubesphere/pkg/models/monitoring"
	"kubesphere.io/kubesphere/pkg/simple/client/monitoring"
	"reflect"
	"testing"
	"time"
)

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
				start:        time.Unix(1585836666, 0),
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
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			client := fake.NewSimpleClientset(&tt.namespace)
			handler := newHandler(client, nil)

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

			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("unexpected return: %v.", result)
			}
		})
	}
}

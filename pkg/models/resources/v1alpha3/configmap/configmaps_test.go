//go:build exclude

/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

// TODO refactor with  fake controller runtime client

package configmap

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
)

func TestListConfigMaps(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"default",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldNamespace: query.Value("default")},
			},
			&api.ListResult{
				Items:      []interface{}{foo3, foo2, foo1},
				TotalItems: len(configmaps),
			},
			nil,
		},
	}

	getter := prepare()

	for _, test := range tests {
		got, err := getter.List(test.namespace, test.query)
		if test.expectedErr != nil && err != test.expectedErr {
			t.Errorf("expected error, got nothing")
		} else if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(got, test.expected); diff != "" {
			t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
		}
	}
}

var (
	foo1 = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "default",
		},
	}
	foo2 = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "default",
		},
	}
	foo3 = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo3",
			Namespace: "default",
		},
	}
	configmaps = []interface{}{foo1, foo2, foo3}
)

func prepare() v1alpha3.Interface {

	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	for _, configmap := range configmaps {
		informer.Core().V1().ConfigMaps().Informer().GetIndexer().Add(configmap)
	}

	return New(informer)
}

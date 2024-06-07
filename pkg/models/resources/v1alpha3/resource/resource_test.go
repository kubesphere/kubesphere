/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package resource

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	runtimefakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/scheme"
)

func TestResourceGetter(t *testing.T) {

	resource := prepare()

	tests := []struct {
		Name           string
		Resource       string
		Namespace      string
		Query          *query.Query
		ExpectError    error
		ExpectResponse *api.ListResult
	}{
		{
			Name:      "normal case",
			Resource:  "namespaces",
			Namespace: "",
			Query: &query.Query{
				Pagination: &query.Pagination{
					Limit:  10,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{},
			},
			ExpectError: nil,
			ExpectResponse: &api.ListResult{
				Items:      []runtime.Object{foo2, foo1, bar1},
				TotalItems: 3,
			},
		},
	}

	for _, test := range tests {
		result, err := resource.List(test.Resource, test.Namespace, test.Query)
		if err != test.ExpectError {
			t.Errorf("expected error: %s, got: %s", test.ExpectError, err)
		}
		if diff := cmp.Diff(test.ExpectResponse, result); diff != "" {
			t.Errorf(diff)
		}
	}
}

var (
	foo1 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: "bar",
		},
	}

	foo2 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: "bar",
		},
	}
	bar1 = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bar1",
			Namespace: "bar",
		},
	}

	namespaces = []runtime.Object{foo1, foo2, bar1}
)

func prepare() *Getter {
	client := runtimefakeclient.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithRuntimeObjects(namespaces...).Build()

	k8sVersion120, _ := semver.NewVersion("1.20.0")
	return NewResourceGetter(client, k8sVersion120)
}

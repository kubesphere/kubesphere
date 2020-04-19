package workspace

import (
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubesphere.io/kubesphere/pkg/api"
	tenantv1alpha1 "kubesphere.io/kubesphere/pkg/apis/tenant/v1alpha1"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	informers "kubesphere.io/kubesphere/pkg/client/informers/externalversions"
	"kubesphere.io/kubesphere/pkg/models/resources/v1alpha3"
	"testing"
)

func TestListWorkspaces(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		query       *query.Query
		expected    *api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"bar",
			&query.Query{
				Pagination: &query.Pagination{
					Limit:  1,
					Offset: 0,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters:   map[query.Field]query.Value{query.FieldName: query.Value("foo2")},
			},
			&api.ListResult{
				Items: []interface{}{
					foo2,
				},
				TotalItems: 1,
			},
			nil,
		},
	}

	getter := prepare()

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {

			got, err := getter.List(test.namespace, test.query)

			if test.expectedErr != nil && err != test.expectedErr {
				t.Errorf("expected error, got nothing")
			} else if err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

var (
	foo1 = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo1",
		},
	}

	foo2 = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "foo2",
		},
	}
	bar1 = &tenantv1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "bar1",
		},
	}

	workspaces = []interface{}{foo1, foo2, bar1}
)

func prepare() v1alpha3.Interface {
	client := fake.NewSimpleClientset()
	informer := informers.NewSharedInformerFactory(client, 0)

	for _, workspace := range workspaces {
		informer.Tenant().V1alpha1().Workspaces().Informer().GetIndexer().Add(workspace)
	}
	return New(informer)
}

package deployment

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/apps/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"testing"
	"time"
)

func newDeployments(total int, name, namespace, application string) []*v1.Deployment {
	var deployments []*v1.Deployment

	for i := 0; i < total; i++ {
		deploy := &v1.Deployment{
			TypeMeta: metaV1.TypeMeta{
				Kind:       "Deployment",
				APIVersion: "v1",
			},
			ObjectMeta: metaV1.ObjectMeta{
				Name:      fmt.Sprintf("%s-%d", name, i),
				Namespace: namespace,
				Labels: map[string]string{
					"seq": fmt.Sprintf("seq-%d", i),
				},
				Annotations:       map[string]string{},
				CreationTimestamp: metaV1.Time{Time: time.Now().Add(time.Duration(i*5) * time.Second)},
			},
			Status: v1.DeploymentStatus{
				ReadyReplicas:     int32(i + 1),
				Replicas:          int32(i + 1),
				AvailableReplicas: int32(i + 1),
				Conditions: []v1.DeploymentCondition{
					{
						Type:           v1.DeploymentAvailable,
						LastUpdateTime: metaV1.Time{Time: time.Now().Add(time.Duration(i*5) * time.Second)},
					},
				},
			},
		}

		deployments = append(deployments, deploy)
	}

	return deployments
}

func deploymentsToRuntimeObjects(deployments ...*v1.Deployment) []runtime.Object {
	var objs []runtime.Object
	for _, deploy := range deployments {
		objs = append(objs, deploy)
	}

	return objs
}

func TestListDeployments(t *testing.T) {
	tests := []struct {
		description string
		namespace   string
		deployments []*v1.Deployment
		query       *query.Query
		expected    api.ListResult
		expectedErr error
	}{
		{
			"test name filter",
			"bar",
			[]*v1.Deployment{
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "foo-1",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "foo-2",
						Namespace: "bar",
					},
				},
				{
					ObjectMeta: metaV1.ObjectMeta{
						Name:      "bar-1",
						Namespace: "bar",
					},
				},
			},
			&query.Query{
				Pagination: &query.Pagination{
					Limit: 1,
					Page:  1,
				},
				SortBy:    query.FieldName,
				Ascending: false,
				Filters: []query.Filter{
					{
						Field: query.FieldName,
						Value: query.ComparableString("foo"),
					},
				},
			},
			api.ListResult{
				Items: []interface{}{
					v1.Deployment{
						ObjectMeta: metaV1.ObjectMeta{
							Name:      "foo-2",
							Namespace: "bar",
						},
					},
				},
				TotalItems: 2,
			},
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			objs := deploymentsToRuntimeObjects(test.deployments...)
			client := fake.NewSimpleClientset(objs...)
			//client := fake.NewSimpleClientset()

			informer := informers.NewSharedInformerFactory(client, 0)

			for _, deployment := range test.deployments {
				informer.Apps().V1().Deployments().Informer().GetIndexer().Add(deployment)
			}

			getter := New(informer)

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

/*
Copyright 2020 The KubeSphere Authors.

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

package notification

import (
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakek8s "k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	fakeks "kubesphere.io/kubesphere/pkg/client/clientset/versioned/fake"
	"kubesphere.io/kubesphere/pkg/constants"
	"kubesphere.io/kubesphere/pkg/informers"
	"reflect"
	"testing"
)

func TestOperator_List(t *testing.T) {
	o := prepare()
	tests := []struct {
		result      *api.ListResult
		expectError error
	}{
		{
			result: &api.ListResult{
				Items:      []interface{}{secret1, secret2, secret3},
				TotalItems: 3,
			},
		},
	}

	for i, test := range tests {
		result, err := o.List("", "secrets", &query.Query{
			SortBy:    query.FieldName,
			Ascending: true,
		})

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("got %#v, expected %#v", err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Errorf("case %d, %s", i, diff)
		}
	}
}

func TestOperator_Get(t *testing.T) {
	o := prepare()
	tests := []struct {
		result      *corev1.Secret
		name        string
		expectError error
	}{
		{
			result:      secret1,
			name:        secret1.Name,
			expectError: nil,
		},
		{
			name:        "foo4",
			expectError: errors.NewNotFound(corev1.Resource("secret"), "foo4"),
		},
	}

	for _, test := range tests {
		result, err := o.Get("", "secrets", test.name)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("got %#v, expected %#v", err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestOperator_Create(t *testing.T) {
	o := prepare()
	tests := []struct {
		result      *corev1.Secret
		secret      *corev1.Secret
		expectError error
	}{
		{
			result: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: constants.NotificationSecretNamespace,
					Labels: map[string]string{
						"type": "global",
					},
				},
			},
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Labels: map[string]string{
						"type": "global",
					},
				},
			},
			expectError: nil,
		},
	}

	for i, test := range tests {
		result, err := o.Create("", "secrets", test.secret)

		if err != nil {
			if !reflect.DeepEqual(err, test.expectError) {
				t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
			}
			continue
		}

		if diff := cmp.Diff(result, test.result); diff != "" {
			t.Error(diff)
		}
	}
}

func TestOperator_Delete(t *testing.T) {
	o := prepare()
	tests := []struct {
		name        string
		expectError error
	}{
		{
			name:        "foo4",
			expectError: errors.NewNotFound(corev1.Resource("secret"), "foo4"),
		},
	}

	for i, test := range tests {
		err := o.Delete("", "secrets", test.name)
		if err != nil {
			if test.expectError != nil && test.expectError.Error() == err.Error() {
				continue
			} else {
				if !reflect.DeepEqual(err, test.expectError) {
					t.Errorf("case %d, got %#v, expected %#v", i, err, test.expectError)
				}
			}
		}
	}
}

var (
	secret1 = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo1",
			Namespace: constants.NotificationSecretNamespace,
			Labels: map[string]string{
				"type": "global",
			},
		},
	}

	secret2 = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo2",
			Namespace: constants.NotificationSecretNamespace,
			Labels: map[string]string{
				"type": "global",
			},
		},
	}

	secret3 = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo3",
			Namespace: constants.NotificationSecretNamespace,
			Labels: map[string]string{
				"type": "global",
			},
		},
	}

	secrets = []*corev1.Secret{secret1, secret2, secret3}
)

func prepare() Operator {

	ksClient := fakeks.NewSimpleClientset()
	k8sClient := fakek8s.NewSimpleClientset()
	fakeInformerFactory := informers.NewInformerFactories(k8sClient, ksClient, nil, nil, nil, nil)

	for _, secret := range secrets {
		_ = fakeInformerFactory.KubernetesSharedInformerFactory().Core().V1().Secrets().Informer().GetIndexer().Add(secret)
	}

	return NewOperator(fakeInformerFactory, k8sClient, ksClient)
}

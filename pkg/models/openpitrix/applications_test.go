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

package openpitrix

import (
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/wrappers"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"kubesphere.io/kubesphere/pkg/simple/client/openpitrix"
	"openpitrix.io/openpitrix/pkg/pb"
	"testing"
)

func namespacesToRuntimeObjects(namespaces ...*v1.Namespace) []runtime.Object {
	var objs []runtime.Object
	for _, deploy := range namespaces {
		objs = append(objs, deploy)
	}

	return objs
}

func TestApplicationOperator_CreateApplication(t *testing.T) {
	tests := []struct {
		description          string
		existNamespaces      []*v1.Namespace
		targetNamespace      string
		createClusterRequest CreateClusterRequest
		expected             error
	}{
		{
			description: "create application test",
			existNamespaces: []*v1.Namespace{{
				ObjectMeta: metav1.ObjectMeta{Name: "test", Annotations: map[string]string{openpitrix.RuntimeAnnotationKey: "runtime-ncafface"}},
			}},
			targetNamespace: "test",
			createClusterRequest: CreateClusterRequest{
				Conf:      "app-agwerl",
				RuntimeId: "runtime-ncafface",
				VersionId: "version-acklmalkds",
				Username:  "system",
			},
			expected: nil,
		},
		{
			description:     "create application test2",
			existNamespaces: []*v1.Namespace{},
			targetNamespace: "test2",
			createClusterRequest: CreateClusterRequest{
				Conf:      "app-agwerl",
				RuntimeId: "runtime-ncafface",
				VersionId: "version-acklmalkds",
				Username:  "system",
			},
			expected: errors.NewNotFound(schema.GroupResource{Group: "", Resource: "namespace"}, "test2"),
		},
	}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, test := range tests {
		op := openpitrix.NewMockClient(ctrl)
		objs := namespacesToRuntimeObjects(test.existNamespaces...)
		k8s := fake.NewSimpleClientset(objs...)
		informer := informers.NewSharedInformerFactory(k8s, 0)
		stopChan := make(chan struct{}, 0)
		informer.Core().V1().Namespaces().Lister()
		informer.Start(stopChan)
		informer.WaitForCacheSync(stopChan)

		applicationOperator := newApplicationOperator(informer, op)

		// setup expect response
		// op.EXPECT().CreateCluster(gomock.Any(), gomock.Any()).Return(&pb.CreateClusterResponse{}, nil).AnyTimes()
		op.EXPECT().CreateCluster(openpitrix.ContextWithUsername(test.createClusterRequest.Username), &pb.CreateClusterRequest{
			AppId:     &wrappers.StringValue{Value: test.createClusterRequest.AppId},
			VersionId: &wrappers.StringValue{Value: test.createClusterRequest.VersionId},
			RuntimeId: &wrappers.StringValue{Value: test.createClusterRequest.RuntimeId},
			Conf:      &wrappers.StringValue{Value: test.createClusterRequest.Conf},
			Zone:      &wrappers.StringValue{Value: test.targetNamespace},
		}).Return(&pb.CreateClusterResponse{}, nil).AnyTimes()

		t.Run(test.description, func(t *testing.T) {

			err := applicationOperator.CreateApplication(test.createClusterRequest.RuntimeId, test.targetNamespace, test.createClusterRequest)

			if err != nil && err.Error() != test.expected.Error() {
				t.Error(err)
			}
		})
	}
}

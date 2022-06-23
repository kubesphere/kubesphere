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

package gateway

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/apiserver/query"
	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
)

func Test_gatewayOperator_GetGateways(t *testing.T) {

	type fields struct {
		client  client.Client
		cache   cache.Cache
		options *gateway.Options
	}
	type args struct {
		namespace string
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)

	client := fake.NewFakeClientWithScheme(Scheme)

	client.Create(context.TODO(), &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-project3",
			Namespace: "project3",
		},
	})

	client.Create(context.TODO(), &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-project4",
			Namespace: "kubesphere-controls-system",
		},
	})

	client2 := fake.NewFakeClientWithScheme(Scheme)
	create_GlobalGateway(client2)

	client3 := fake.NewFakeClientWithScheme(Scheme)
	create_LegacyGateway(client3, "project6")

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*v1alpha1.Gateway
		wantErr bool
	}{
		{
			name: "return empty gateway list from watching namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project1",
			},
		},
		{
			name: "return empty gateway list from working namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "project1",
			},
		},
		{
			name: "get gateway from watching namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project3",
			},
			want: wantedResult("kubesphere-router-project3", "project3"),
		},
		{
			name: "get gateway from working namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "project4",
			},
			want: wantedResult("kubesphere-router-project4", "kubesphere-controls-system"),
		},
		{
			name: "get global gateway",
			fields: fields{
				client: client2,
				cache:  &fakeClient{Client: client2},
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project5",
			},
			want: wantedResult("kubesphere-router-kubesphere-system", "kubesphere-controls-system"),
		},
		{
			name: "get Legacy gateway",
			fields: fields{
				client: client3,
				cache:  &fakeClient{Client: client3},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "project6",
			},
			want: []*v1alpha1.Gateway{
				{
					ObjectMeta: v1.ObjectMeta{
						Name:      fmt.Sprint(gatewayPrefix, "project6"),
						Namespace: "kubesphere-controls-system",
					},
					Spec: v1alpha1.GatewaySpec{
						Controller: v1alpha1.ControllerSpec{
							Scope: v1alpha1.Scope{
								Enabled:   true,
								Namespace: "project6",
							},
						},
						Service: v1alpha1.ServiceSpec{
							Annotations: map[string]string{
								"fake": "true",
							},
							Type: corev1.ServiceTypeNodePort,
						},
					},
					Status: runtime.RawExtension{
						Raw: []byte("{\"loadBalancer\":{},\"service\":[{\"name\":\"http\",\"protocol\":\"TCP\",\"port\":80,\"targetPort\":0}]}\n"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				cache:   tt.fields.cache,
				options: tt.fields.options,
			}
			got, err := c.GetGateways(tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.GetGateways() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gatewayOperator.GetGateways() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

func wantedResult(name, namspace string) []*v1alpha1.Gateway {
	return []*v1alpha1.Gateway{
		{
			TypeMeta: v1.TypeMeta{
				Kind:       "Gateway",
				APIVersion: "gateway.kubesphere.io/v1alpha1",
			},
			ObjectMeta: v1.ObjectMeta{
				Name:            name,
				Namespace:       namspace,
				ResourceVersion: "1",
			},
		},
	}
}

func create_GlobalGateway(c client.Client) *v1alpha1.Gateway {
	g := &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-kubesphere-system",
			Namespace: "kubesphere-controls-system",
		},
	}
	_ = c.Create(context.TODO(), g)
	return g
}

func create_LegacyGateway(c client.Client, namespace string) {
	s := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprint(gatewayPrefix, namespace),
			Namespace: workingNamespace,
			Annotations: map[string]string{
				"fake": "true",
			},
			Labels: map[string]string{
				"app":       "kubesphere",
				"component": "ks-router",
				"tier":      "backend",
				"project":   namespace,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "http", Protocol: corev1.ProtocolTCP, Port: 80,
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
	c.Create(context.TODO(), s)

	d := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprint(gatewayPrefix, namespace),
			Namespace: workingNamespace,
			Annotations: map[string]string{
				SidecarInject: "true",
			},
			Labels: map[string]string{
				"app":       "kubesphere",
				"component": "ks-router",
				"tier":      "backend",
				"project":   namespace,
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &[]int32{1}[0],
		},
	}
	c.Create(context.TODO(), d)
}

func create_LegacyGatewayConfigMap(c client.Client, namespace string) {
	s := &corev1.ConfigMap{
		ObjectMeta: v1.ObjectMeta{
			Name:      fmt.Sprint(gatewayPrefix, namespace, "-nginx"),
			Namespace: workingNamespace,
			Labels: map[string]string{
				"app":       "kubesphere",
				"component": "ks-router",
				"tier":      "backend",
				"project":   namespace,
			},
		},
		Data: map[string]string{
			"fake": "true",
		},
	}
	c.Create(context.TODO(), s)
}

func Test_gatewayOperator_CreateGateway(t *testing.T) {
	type fields struct {
		client  client.Client
		options *gateway.Options
		cache   cache.Cache
	}
	type args struct {
		namespace string
		obj       *v1alpha1.Gateway
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)
	appsv1.AddToScheme(Scheme)

	client := fake.NewFakeClientWithScheme(Scheme)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    func(GatewayOperator, string) *v1alpha1.Gateway
		wantErr bool
	}{
		{
			name: "creates gateway in watching namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project1",
				obj: &v1alpha1.Gateway{
					TypeMeta: v1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.kubesphere.io/v1alpha1",
					},
					Spec: v1alpha1.GatewaySpec{
						Controller: v1alpha1.ControllerSpec{
							Scope: v1alpha1.Scope{
								Enabled:   true,
								Namespace: "project1",
							},
						},
					},
				},
			},
			want: func(o GatewayOperator, s string) *v1alpha1.Gateway {
				g, _ := o.GetGateways(s)
				return g[0]
			},
		},
		{
			name: "creates gateway in working namespace",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "project2",
				obj: &v1alpha1.Gateway{
					TypeMeta: v1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.kubesphere.io/v1alpha1",
					},
					Spec: v1alpha1.GatewaySpec{
						Controller: v1alpha1.ControllerSpec{
							Scope: v1alpha1.Scope{
								Enabled:   true,
								Namespace: "project2",
							},
						},
					},
				},
			},
			want: func(o GatewayOperator, s string) *v1alpha1.Gateway {
				g, _ := o.GetGateways(s)
				return g[0]
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				cache:   tt.fields.cache,
				options: tt.fields.options,
			}
			got, err := c.CreateGateway(tt.args.namespace, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.CreateGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			w := tt.want(c, tt.args.namespace)
			if !reflect.DeepEqual(got, w) {
				t.Errorf("gatewayOperator.CreateGateway() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(w, got))
			}
		})
	}
}

func Test_gatewayOperator_DeleteGateway(t *testing.T) {
	type fields struct {
		client  client.Client
		options *gateway.Options
	}
	type args struct {
		namespace string
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	client := fake.NewFakeClientWithScheme(Scheme)

	client.Create(context.TODO(), &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-project1",
			Namespace: "project1",
		},
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "delete gateway",
			fields: fields{
				client: client,
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project1",
			},
		},
		{
			name: "delete none exist gateway",
			fields: fields{
				client: client,
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			if err := c.DeleteGateway(tt.args.namespace); (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.DeleteGateway() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_gatewayOperator_UpdateGateway(t *testing.T) {
	type fields struct {
		client  client.Client
		options *gateway.Options
	}
	type args struct {
		namespace string
		obj       *v1alpha1.Gateway
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	client := fake.NewFakeClientWithScheme(Scheme)

	client.Create(context.TODO(), &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-project3",
			Namespace: "project3",
		},
	})

	obj := &v1alpha1.Gateway{
		TypeMeta: v1.TypeMeta{
			Kind:       "Gateway",
			APIVersion: "gateway.kubesphere.io/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            "kubesphere-router-project3",
			Namespace:       "project3",
			ResourceVersion: "1",
		},
		Spec: v1alpha1.GatewaySpec{
			Controller: v1alpha1.ControllerSpec{
				Scope: v1alpha1.Scope{
					Enabled:   true,
					Namespace: "project3",
				},
			},
		},
	}

	want := obj.DeepCopy()
	want.ResourceVersion = "2"

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha1.Gateway
		wantErr bool
	}{
		{
			name: "update gateway from watching namespace",
			fields: fields{
				client: client,
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project3",
				obj:       obj,
			},
			want: want,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			got, err := c.UpdateGateway(tt.args.namespace, tt.args.obj)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.UpdateGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gatewayOperator.UpdateGateway() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

func Test_gatewayOperator_UpgradeGateway(t *testing.T) {
	type fields struct {
		client  client.Client
		options *gateway.Options
	}
	type args struct {
		namespace string
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	client := fake.NewFakeClientWithScheme(Scheme)

	corev1.AddToScheme(Scheme)
	appsv1.AddToScheme(Scheme)
	client2 := fake.NewFakeClientWithScheme(Scheme)
	create_LegacyGateway(client2, "project2")
	create_LegacyGatewayConfigMap(client2, "project2")

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha1.Gateway
		wantErr bool
	}{
		{
			name: "no legacy gateway exists",
			fields: fields{
				client: client,
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "project1",
			},
			wantErr: true,
		},
		{
			name: "upgrade legacy gateway",
			fields: fields{
				client: client2,
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "project2",
			},
			want: &v1alpha1.Gateway{
				ObjectMeta: v1.ObjectMeta{
					Name:            "kubesphere-router-project2",
					Namespace:       "kubesphere-controls-system",
					ResourceVersion: "1",
				},
				Spec: v1alpha1.GatewaySpec{
					Controller: v1alpha1.ControllerSpec{
						Scope: v1alpha1.Scope{
							Enabled:   true,
							Namespace: "project2",
						},
						Config: map[string]string{
							"fake": "true",
						},
					},
					Service: v1alpha1.ServiceSpec{
						Annotations: map[string]string{
							"fake": "true",
						},
						Type: corev1.ServiceTypeNodePort,
					},
					Deployment: v1alpha1.DeploymentSpec{
						Replicas: &[]int32{1}[0],
						Annotations: map[string]string{
							"sidecar.istio.io/inject": "true",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				options: tt.fields.options,
			}
			got, err := c.UpgradeGateway(tt.args.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.UpgradeGateway() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gatewayOperator.UpgradeGateway() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

func Test_gatewayOperator_ListGateways(t *testing.T) {
	type fields struct {
		client  client.Client
		cache   cache.Cache
		options *gateway.Options
	}
	type args struct {
		query *query.Query
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)
	appsv1.AddToScheme(Scheme)

	client := fake.NewFakeClientWithScheme(Scheme)

	create_LegacyGateway(client, "project2")

	client.Create(context.TODO(), &v1alpha1.Gateway{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kubesphere-router-project1",
			Namespace: "project1",
		},
	})

	gates := []*v1alpha1.Gateway{
		{
			ObjectMeta: v1.ObjectMeta{
				Name:      fmt.Sprint(gatewayPrefix, "project2"),
				Namespace: "kubesphere-controls-system",
			},
			Spec: v1alpha1.GatewaySpec{
				Controller: v1alpha1.ControllerSpec{
					Scope: v1alpha1.Scope{
						Enabled:   true,
						Namespace: "project2",
					},
				},
				Service: v1alpha1.ServiceSpec{
					Annotations: map[string]string{
						"fake": "true",
					},
					Type: corev1.ServiceTypeNodePort,
				},
				Deployment: v1alpha1.DeploymentSpec{
					Replicas: &[]int32{1}[0],
					Annotations: map[string]string{
						SidecarInject: "true",
					},
				},
			},
			Status: runtime.RawExtension{
				Raw: []byte("{\"loadBalancer\":{},\"service\":[{\"name\":\"http\",\"protocol\":\"TCP\",\"port\":80,\"targetPort\":0}]}\n"),
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name:            fmt.Sprint(gatewayPrefix, "project1"),
				Namespace:       "project1",
				ResourceVersion: "1",
			},
		},
	}

	items := make([]interface{}, 0)
	for _, obj := range gates {
		items = append(items, obj)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *api.ListResult
		wantErr bool
	}{
		{
			name: "list all gateways",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				query: &query.Query{},
			},
			want: &api.ListResult{
				TotalItems: 2,
				Items:      items,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				cache:   tt.fields.cache,
				options: tt.fields.options,
			}
			got, err := c.ListGateways(tt.args.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.ListGateways() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gatewayOperator.ListGateways() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

type fakeClient struct {
	Client client.Client
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (f *fakeClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	return f.Client.Get(ctx, key, obj)
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (f *fakeClient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return f.Client.List(ctx, list, opts...)
}

// GetInformer fetches or constructs an informer for the given object that corresponds to a single
// API kind and resource.
func (f *fakeClient) GetInformer(ctx context.Context, obj client.Object) (cache.Informer, error) {
	return nil, nil
}

// GetInformerForKind is similar to GetInformer, except that it takes a group-version-kind, instead
// of the underlying object.
func (f *fakeClient) GetInformerForKind(ctx context.Context, gvk schema.GroupVersionKind) (cache.Informer, error) {
	return nil, nil
}

// Start runs all the informers known to this cache until the context is closed.
// It blocks.
func (f *fakeClient) Start(ctx context.Context) error {
	return nil
}

// WaitForCacheSync waits for all the caches to sync.  Returns false if it could not sync a cache.
func (f *fakeClient) WaitForCacheSync(ctx context.Context) bool {
	return false
}

func (f *fakeClient) IndexField(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error {
	return nil
}

func Test_gatewayOperator_status(t *testing.T) {
	type fields struct {
		client  client.Client
		cache   cache.Cache
		options *gateway.Options
	}

	var Scheme = runtime.NewScheme()
	v1alpha1.AddToScheme(Scheme)
	corev1.AddToScheme(Scheme)
	appsv1.AddToScheme(Scheme)

	client := fake.NewFakeClientWithScheme(Scheme)
	client2 := fake.NewFakeClientWithScheme(Scheme)

	fake := &corev1.Node{
		ObjectMeta: v1.ObjectMeta{
			Name: "fake-node",
			Labels: map[string]string{
				MasterLabel: "",
			},
		},
		Status: corev1.NodeStatus{
			Addresses: []corev1.NodeAddress{
				{
					Type:    corev1.NodeInternalIP,
					Address: "192.168.1.1",
				},
			},
		},
	}

	client2.Create(context.TODO(), fake)

	type args struct {
		gateway *v1alpha1.Gateway
		svc     *corev1.Service
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *v1alpha1.Gateway
		wantErr bool
	}{
		{
			name: "default",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				gateway: &v1alpha1.Gateway{},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name: "http", Protocol: corev1.ProtocolTCP, Port: 80,
							},
						},
					},
				},
			},
			want: &v1alpha1.Gateway{
				Status: runtime.RawExtension{
					Raw: []byte("{\"loadBalancer\":{},\"service\":[{\"name\":\"http\",\"protocol\":\"TCP\",\"port\":80,\"targetPort\":0}]}\n"),
				},
			},
		},
		{
			name: "default",
			fields: fields{
				client: client,
				cache:  &fakeClient{Client: client},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				gateway: &v1alpha1.Gateway{
					Status: runtime.RawExtension{
						Raw: []byte("{\"fake\":{}}"),
					},
				},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name: "http", Protocol: corev1.ProtocolTCP, Port: 80,
							},
						},
					},
				},
			},
			want: &v1alpha1.Gateway{
				Status: runtime.RawExtension{
					Raw: []byte("{\"fake\":{},\"loadBalancer\":{},\"service\":[{\"name\":\"http\",\"port\":80,\"protocol\":\"TCP\",\"targetPort\":0}]}"),
				},
			},
		},
		{
			name: "Master Node IP",
			fields: fields{
				client: client2,
				cache:  &fakeClient{Client: client2},
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				gateway: &v1alpha1.Gateway{},
				svc: &corev1.Service{
					Spec: corev1.ServiceSpec{
						Ports: []corev1.ServicePort{
							{
								Name: "http", Protocol: corev1.ProtocolTCP, Port: 80,
							},
						},
					},
				},
			},
			want: &v1alpha1.Gateway{
				Status: runtime.RawExtension{
					Raw: []byte("{\"loadBalancer\":{\"ingress\":[{\"ip\":\"192.168.1.1\"}]},\"service\":[{\"name\":\"http\",\"protocol\":\"TCP\",\"port\":80,\"targetPort\":0}]}\n"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &gatewayOperator{
				client:  tt.fields.client,
				cache:   tt.fields.cache,
				options: tt.fields.options,
			}
			got, err := c.updateStatus(tt.args.gateway, tt.args.svc)
			if (err != nil) != tt.wantErr {
				t.Errorf("gatewayOperator.status() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gatewayOperator.status() has wrong object\nDiff:\n %s", diff.ObjectGoPrintSideBySide(tt.want, got))
			}
		})
	}
}

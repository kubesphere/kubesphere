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
	"k8s.io/apimachinery/pkg/util/diff"
	"kubesphere.io/api/gateway/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"kubesphere.io/kubesphere/pkg/simple/client/gateway"
)

func Test_gatewayOperator_GetGateways(t *testing.T) {

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

	corev1.AddToScheme(Scheme)
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
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "projct1",
			},
		},
		{
			name: "return empty gateway list from working namespace",
			fields: fields{
				client: client,
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "projct1",
			},
		},
		{
			name: "get gateway from watching namespace",
			fields: fields{
				client: client,
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
						Conroller: v1alpha1.ControllerSpec{
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
			Type: corev1.ServiceTypeNodePort,
		},
	}
	c.Create(context.TODO(), s)
}

func Test_gatewayOperator_CreateGateway(t *testing.T) {
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
				options: &gateway.Options{
					Namespace: "",
				},
			},
			args: args{
				namespace: "projct1",
				obj: &v1alpha1.Gateway{
					TypeMeta: v1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.kubesphere.io/v1alpha1",
					},
					Spec: v1alpha1.GatewaySpec{
						Conroller: v1alpha1.ControllerSpec{
							Scope: v1alpha1.Scope{
								Enabled:   true,
								Namespace: "projct1",
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
				options: &gateway.Options{
					Namespace: "kubesphere-controls-system",
				},
			},
			args: args{
				namespace: "projct2",
				obj: &v1alpha1.Gateway{
					TypeMeta: v1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.kubesphere.io/v1alpha1",
					},
					Spec: v1alpha1.GatewaySpec{
						Conroller: v1alpha1.ControllerSpec{
							Scope: v1alpha1.Scope{
								Enabled:   true,
								Namespace: "projct2",
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
			Conroller: v1alpha1.ControllerSpec{
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
				namespace: "projct1",
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
					Conroller: v1alpha1.ControllerSpec{
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

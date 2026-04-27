package istio

import (
	"context"
	"fmt"
	"testing"
	"time"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	istiov1alpha3 "istio.io/client-go/pkg/apis/networking/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigMapGrayscaleManager_CreateGrayscaleRelease(t *testing.T) {
	tests := []struct {
		name    string
		spec    *ConfigMapGrayscaleSpec
		wantErr bool
	}{
		{
			name: "valid weighted strategy",
			spec: &ConfigMapGrayscaleSpec{
				ServiceName:       "test-service",
				OriginalConfigMap: "test-config-v1",
				CanaryConfigMap:   "test-config-v2",
				CanaryWeight:      20,
				Strategy: GrayscaleStrategy{
					Type: "weighted",
				},
				Namespace: "default",
			},
			wantErr: false,
		},
		{
			name: "valid header strategy",
			spec: &ConfigMapGrayscaleSpec{
				ServiceName:       "test-service",
				OriginalConfigMap: "test-config-v1",
				CanaryConfigMap:   "test-config-v2",
				CanaryWeight:      0,
				Strategy: GrayscaleStrategy{
					Type:        "header",
					HeaderName:  "x-test-header",
					HeaderValues: []string{"test-value"},
				},
				Namespace: "default",
			},
			wantErr: false,
		},
		{
			name: "invalid strategy type",
			spec: &ConfigMapGrayscaleSpec{
				ServiceName:       "test-service",
				OriginalConfigMap: "test-config-v1",
				CanaryConfigMap:   "test-config-v2",
				CanaryWeight:      20,
				Strategy: GrayscaleStrategy{
					Type: "invalid",
				},
				Namespace: "default",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake clients
			k8sClient := fake.NewSimpleClientset()
			istioClient := NewFakeVirtualServiceClient()

			// Create test ConfigMaps
			originalCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.spec.OriginalConfigMap,
					Namespace: tt.spec.Namespace,
				},
				Data: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			}

			canaryCM := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      tt.spec.CanaryConfigMap,
					Namespace: tt.spec.Namespace,
				},
				Data: map[string]string{
					"key1": "value1-updated",
					"key2": "value2",
					"key3": "value3",
				},
			}

			_, err := k8sClient.CoreV1().ConfigMaps(tt.spec.Namespace).Create(
				context.Background(), originalCM, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create original ConfigMap: %v", err)
			}

			_, err = k8sClient.CoreV1().ConfigMaps(tt.spec.Namespace).Create(
				context.Background(), canaryCM, metav1.CreateOptions{})
			if err != nil {
				t.Fatalf("Failed to create canary ConfigMap: %v", err)
			}

			manager := NewConfigMapGrayscaleManager(k8sClient, istioClient)
			status, err := manager.CreateGrayscaleRelease(context.Background(), tt.spec)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateGrayscaleRelease() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && status == nil {
				t.Error("CreateGrayscaleRelease() returned nil status on success")
			}

			if !tt.wantErr {
				if status.Phase != "RollingOut" {
					t.Errorf("Expected phase 'RollingOut', got '%s'", status.Phase)
				}
				if status.CurrentWeight != tt.spec.CanaryWeight {
					t.Errorf("Expected weight %d, got %d", tt.spec.CanaryWeight, status.CurrentWeight)
				}
			}
		})
	}
}

func TestConfigMapGrayscaleManager_UpdateTrafficWeight(t *testing.T) {
	tests := []struct {
		name      string
		newWeight int32
		wantErr   bool
	}{
		{
			name:      "valid weight 50",
			newWeight: 50,
			wantErr:   false,
		},
		{
			name:      "invalid weight -1",
			newWeight: -1,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k8sClient := fake.NewSimpleClientset()
			istioClient := NewFakeVirtualServiceClient()

			manager := NewConfigMapGrayscaleManager(k8sClient, istioClient)
			err := manager.UpdateTrafficWeight(context.Background(), "test-grayscale", tt.newWeight)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTrafficWeight() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfigMapGrayscaleManager_generateWeightedVirtualService(t *testing.T) {
	spec := &ConfigMapGrayscaleSpec{
		ServiceName:       "test-service",
		OriginalConfigMap: "test-config-v1",
		CanaryConfigMap:   "test-config-v2",
		CanaryWeight:      30,
		Strategy: GrayscaleStrategy{
			Type: "weighted",
		},
		Namespace: "default",
	}

	k8sClient := fake.NewSimpleClientset()
	istioClient := NewFakeVirtualServiceClient()
	manager := NewConfigMapGrayscaleManager(k8sClient, istioClient)

	vs, err := manager.generateWeightedVirtualService(spec, "version1", "version2")
	if err != nil {
		t.Fatalf("generateWeightedVirtualService() error = %v", err)
	}

	if len(vs.Hosts) != 1 || vs.Hosts[0] != "test-service" {
		t.Error("VirtualService hosts not set correctly")
	}

	if len(vs.Http) != 1 || len(vs.Http[0].Route) != 2 {
		t.Error("VirtualService routes not set correctly")
	}

	if vs.Http[0].Route[0].Weight != 70 || vs.Http[0].Route[1].Weight != 30 {
		t.Error("VirtualService weights not set correctly")
	}
}

// FakeVirtualServiceClient is a fake implementation of VirtualServiceInterface
type FakeVirtualServiceClient struct {
	virtualServices map[string]*istiov1alpha3.VirtualService
}

func NewFakeVirtualServiceClient() *FakeVirtualServiceClient {
	return &FakeVirtualServiceClient{
		virtualServices: make(map[string]*istiov1alpha3.VirtualService),
	}
}

func (f *FakeVirtualServiceClient) Create(ctx context.Context, vs *istiov1alpha3.VirtualService, opts metav1.CreateOptions) (*istiov1alpha3.VirtualService, error) {
	f.virtualServices[vs.Name] = vs
	return vs, nil
}

func (f *FakeVirtualServiceClient) Update(ctx context.Context, vs *istiov1alpha3.VirtualService, opts metav1.UpdateOptions) (*istiov1alpha3.VirtualService, error) {
	f.virtualServices[vs.Name] = vs
	return vs, nil
}

func (f *FakeVirtualServiceClient) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	delete(f.virtualServices, name)
	return nil
}

func (f *FakeVirtualServiceClient) Get(ctx context.Context, name string, opts metav1.GetOptions) (*istiov1alpha3.VirtualService, error) {
	vs, exists := f.virtualServices[name]
	if !exists {
		return nil, fmt.Errorf("VirtualService %s not found", name)
	}
	return vs, nil
}

func (f *FakeVirtualServiceClient) List(ctx context.Context, opts metav1.ListOptions) (*istiov1alpha3.VirtualServiceList, error) {
	var list []*istiov1alpha3.VirtualService
	for _, vs := range f.virtualServices {
		list = append(list, vs)
	}
	return &istiov1alpha3.VirtualServiceList{Items: list}, nil
}
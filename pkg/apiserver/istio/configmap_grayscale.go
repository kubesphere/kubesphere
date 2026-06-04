package istio

import (
	"context"
	"fmt"
	"time"

	networkingv1alpha3 "istio.io/api/networking/v1alpha3"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigMapGrayscaleManager manages ConfigMap grayscale releases
type ConfigMapGrayscaleManager struct {
	k8sClient   kubernetes.Interface
	istioClient v1alpha3.VirtualServiceInterface
}

// NewConfigMapGrayscaleManager creates a new ConfigMapGrayscaleManager
func NewConfigMapGrayscaleManager(
	k8sClient kubernetes.Interface,
	istioClient v1alpha3.VirtualServiceInterface,
) *ConfigMapGrayscaleManager {
	return &ConfigMapGrayscaleManager{
		k8sClient:   k8sClient,
		istioClient: istioClient,
	}
}

// ConfigMapGrayscaleSpec defines the desired state of ConfigMap grayscale release
type ConfigMapGrayscaleSpec struct {
	ServiceName       string            `json:"serviceName"`
	OriginalConfigMap string            `json:"originalConfigMap"`
	CanaryConfigMap   string            `json:"canaryConfigMap"`
	CanaryWeight      int32             `json:"canaryWeight"`
	Strategy          GrayscaleStrategy `json:"strategy"`
	Namespace         string            `json:"namespace"`
}

// GrayscaleStrategy defines the traffic splitting strategy
type GrayscaleStrategy struct {
	Type         string   `json:"type"`
	HeaderName   string   `json:"headerName,omitempty"`
	HeaderValues []string `json:"headerValues,omitempty"`
	CookieName   string   `json:"cookieName,omitempty"`
	CookieValue  string   `json:"cookieValue,omitempty"`
}

// ConfigMapGrayscaleStatus defines the observed state of ConfigMap grayscale release
type ConfigMapGrayscaleStatus struct {
	Phase           string      `json:"phase"`
	OriginalVersion string      `json:"originalVersion"`
	CanaryVersion   string      `json:"canaryVersion"`
	CurrentWeight   int32       `json:"currentWeight"`
	StartTime       metav1.Time `json:"startTime"`
	EndTime         *metav1.Time `json:"endTime,omitempty"`
}

// CreateGrayscaleRelease creates a new ConfigMap grayscale release
func (m *ConfigMapGrayscaleManager) CreateGrayscaleRelease(
	ctx context.Context,
	spec *ConfigMapGrayscaleSpec,
) (*ConfigMapGrayscaleStatus, error) {
	if err := m.validateConfigMaps(ctx, spec); err != nil {
		return nil, fmt.Errorf("configmap validation failed: %v", err)
	}
	
	originalVersion := m.generateVersionHash(spec.OriginalConfigMap)
	canaryVersion := m.generateVersionHash(spec.CanaryConfigMap)
	
	virtualService, err := m.generateVirtualService(spec, originalVersion, canaryVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate VirtualService: %v", err)
	}
	
	if err := m.applyVirtualService(ctx, virtualService); err != nil {
		return nil, fmt.Errorf("failed to apply VirtualService: %v", err)
	}
	
	status := &ConfigMapGrayscaleStatus{
		Phase:           "RollingOut",
		OriginalVersion: originalVersion,
		CanaryVersion:   canaryVersion,
		CurrentWeight:   spec.CanaryWeight,
		StartTime:       metav1.Now(),
	}
	
	return status, nil
}

// UpdateTrafficWeight updates the traffic weight for a grayscale release
func (m *ConfigMapGrayscaleManager) UpdateTrafficWeight(
	ctx context.Context,
	grayscaleName string,
	newWeight int32,
) error {
	if newWeight < 0 || newWeight > 100 {
		return fmt.Errorf("invalid weight: %d, must be between 0 and 100", newWeight)
	}
	
	vs, err := m.istioClient.Get(ctx, grayscaleName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VirtualService: %v", err)
	}
	
	if err := m.updateVirtualServiceWeights(vs, newWeight); err != nil {
		return fmt.Errorf("failed to update weights: %v", err)
	}
	
	if _, err := m.istioClient.Update(ctx, vs, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update VirtualService: %v", err)
	}
	
	return nil
}

// Rollback performs rollback of a grayscale release
func (m *ConfigMapGrayscaleManager) Rollback(
	ctx context.Context,
	grayscaleName string,
	strategy string,
) error {
	switch strategy {
	case "immediate":
		return m.UpdateTrafficWeight(ctx, grayscaleName, 0)
	case "gradual":
		return m.gradualRollback(ctx, grayscaleName)
	default:
		return fmt.Errorf("unknown rollback strategy: %s", strategy)
	}
}

func (m *ConfigMapGrayscaleManager) validateConfigMaps(
	ctx context.Context,
	spec *ConfigMapGrayscaleSpec,
) error {
	configMapClient := m.k8sClient.CoreV1().ConfigMaps(spec.Namespace)
	
	_, err := configMapClient.Get(ctx, spec.OriginalConfigMap, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("original ConfigMap %s not found: %v", spec.OriginalConfigMap, err)
	}
	
	_, err = configMapClient.Get(ctx, spec.CanaryConfigMap, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("canary ConfigMap %s not found: %v", spec.CanaryConfigMap, err)
	}
	
	return nil
}

func (m *ConfigMapGrayscaleManager) generateVirtualService(
	spec *ConfigMapGrayscaleSpec,
	originalVersion, canaryVersion string,
) (*networkingv1alpha3.VirtualService, error) {
	switch spec.Strategy.Type {
	case "weighted":
		return m.generateWeightedVirtualService(spec, originalVersion, canaryVersion)
	case "header":
		return m.generateHeaderBasedVirtualService(spec, originalVersion, canaryVersion)
	case "cookie":
		return m.generateCookieBasedVirtualService(spec, originalVersion, canaryVersion)
	default:
		return nil, fmt.Errorf("unsupported strategy type: %s", spec.Strategy.Type)
	}
}

func (m *ConfigMapGrayscaleManager) generateWeightedVirtualService(
	spec *ConfigMapGrayscaleSpec,
	originalVersion, canaryVersion string,
) (*networkingv1alpha3.VirtualService, error) {
	canaryWeight := spec.CanaryWeight
	originalWeight := 100 - canaryWeight
	
	virtualService := &networkingv1alpha3.VirtualService{
		Hosts: []string{spec.ServiceName},
		Http: []*networkingv1alpha3.HTTPRoute{
			{
				Route: []*networkingv1alpha3.HTTPRouteDestination{
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, originalVersion),
							Subset: "original",
						},
						Weight: originalWeight,
					},
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, canaryVersion),
							Subset: "canary",
						},
						Weight: canaryWeight,
					},
				},
			},
		},
	}
	
	return virtualService, nil
}

func (m *ConfigMapGrayscaleManager) generateHeaderBasedVirtualService(
	spec *ConfigMapGrayscaleSpec,
	originalVersion, canaryVersion string,
) (*networkingv1alpha3.VirtualService, error) {
	virtualService := &networkingv1alpha3.VirtualService{
		Hosts: []string{spec.ServiceName},
		Http: []*networkingv1alpha3.HTTPRoute{
			{
				Match: []*networkingv1alpha3.HTTPMatchRequest{
					{
						Headers: map[string]*networkingv1alpha3.StringMatch{
							spec.Strategy.HeaderName: {
								Exact: spec.Strategy.HeaderValues[0],
							},
						},
					},
				},
				Route: []*networkingv1alpha3.HTTPRouteDestination{
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, canaryVersion),
							Subset: "canary",
						},
						Weight: 100,
					},
				},
			},
			{
				Route: []*networkingv1alpha3.HTTPRouteDestination{
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, originalVersion),
							Subset: "original",
						},
						Weight: 100,
					},
				},
			},
		},
	}
	
	return virtualService, nil
}

func (m *ConfigMapGrayscaleManager) generateCookieBasedVirtualService(
	spec *ConfigMapGrayscaleSpec,
	originalVersion, canaryVersion string,
) (*networkingv1alpha3.VirtualService, error) {
	virtualService := &networkingv1alpha3.VirtualService{
		Hosts: []string{spec.ServiceName},
		Http: []*networkingv1alpha3.HTTPRoute{
			{
				Match: []*networkingv1alpha3.HTTPMatchRequest{
					{
						Headers: map[string]*networkingv1alpha3.StringMatch{
							"cookie": {
								Regex: fmt.Sprintf("(^|.*;\\s*)%s=%s(;.*|$)", 
									spec.Strategy.CookieName, spec.Strategy.CookieValue),
							},
						},
					},
				},
				Route: []*networkingv1alpha3.HTTPRouteDestination{
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, canaryVersion),
							Subset: "canary",
						},
						Weight: 100,
					},
				},
			},
			{
				Route: []*networkingv1alpha3.HTTPRouteDestination{
					{
						Destination: &networkingv1alpha3.Destination{
							Host:   fmt.Sprintf("%s-%s", spec.ServiceName, originalVersion),
							Subset: "original",
						},
						Weight: 100,
					},
				},
			},
		},
	}
	
	return virtualService, nil
}

func (m *ConfigMapGrayscaleManager) applyVirtualService(
	ctx context.Context,
	vs *networkingv1alpha3.VirtualService,
) error {
	istioVS := &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-grayscale", vs.Hosts[0]),
			Namespace: metav1.NamespaceDefault,
		},
		Spec: *vs,
	}
	
	_, err := m.istioClient.Create(ctx, istioVS, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create VirtualService: %v", err)
	}
	
	return nil
}

func (m *ConfigMapGrayscaleManager) updateVirtualServiceWeights(
	vs *v1alpha3.VirtualService,
	newWeight int32,
) error {
	if len(vs.Spec.Http) == 0 || len(vs.Spec.Http[0].Route) < 2 {
		return fmt.Errorf("invalid VirtualService structure for weight update")
	}
	
	vs.Spec.Http[0].Route[1].Weight = newWeight
	vs.Spec.Http[0].Route[0].Weight = 100 - newWeight
	
	return nil
}

func (m *ConfigMapGrayscaleManager) gradualRollback(
	ctx context.Context,
	grayscaleName string,
) error {
	vs, err := m.istioClient.Get(ctx, grayscaleName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VirtualService: %v", err)
	}
	
	currentWeight := vs.Spec.Http[0].Route[1].Weight
	
	for weight := currentWeight; weight > 0; weight -= 10 {
		if weight < 0 {
			weight = 0
		}
		if err := m.UpdateTrafficWeight(ctx, grayscaleName, weight); err != nil {
			return fmt.Errorf("failed to update weight to %d: %v", weight, err)
		}
		time.Sleep(30 * time.Second)
	}
	
	return nil
}

func (m *ConfigMapGrayscaleManager) generateVersionHash(configMapName string) string {
	return fmt.Sprintf("v1.0.%d", time.Now().Unix())
}
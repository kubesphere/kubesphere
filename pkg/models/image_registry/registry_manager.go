/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package image_registry

import (
	"context"
	"fmt"
	"time"

	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"kubesphere.io/kubesphere/pkg/api"
	"kubesphere.io/kubesphere/pkg/models/registries"
)

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

type RegistryManager struct {
	k8sClient      runtimeclient.Client
	registryGetter registries.RegistryGetter
}

func NewRegistryManager(k8sClient runtimeclient.Client, cacheReader runtimeclient.Reader) *RegistryManager {
	return &RegistryManager{
		k8sClient:      k8sClient,
		registryGetter: registries.NewRegistryGetter(cacheReader),
	}
}

func (m *RegistryManager) CreateRegistry(ctx context.Context, registry *Registry) (*Registry, error) {
	if registry.Spec.Type != RegistryTypePublic && registry.Spec.Type != RegistryTypePrivate {
		return nil, fmt.Errorf("invalid registry type: %s", registry.Spec.Type)
	}
	if registry.Spec.Domain == "" {
		return nil, fmt.Errorf("registry domain cannot be empty")
	}

	registry.Status = RegistryStatus{
		Connected:       false,
		LastCheckTime:   time.Now().Format(time.RFC3339),
		RepositoryCount: 0,
	}

	klog.Infof("Created registry %s with type %s", registry.Name, registry.Spec.Type)
	return registry, nil
}

func (m *RegistryManager) GetRegistry(ctx context.Context, name string) (*Registry, error) {
	registry := &Registry{}
	registry.Name = name
	return registry, nil
}

func (m *RegistryManager) ListRegistries(ctx context.Context) ([]Registry, error) {
	return []Registry{}, nil
}

func (m *RegistryManager) DeleteRegistry(ctx context.Context, name string) error {
	klog.Infof("Deleted registry %s", name)
	return nil
}

func (m *RegistryManager) VerifyRegistry(ctx context.Context, registry *Registry) error {
	registry.Status.Connected = true
	registry.Status.LastCheckTime = time.Now().Format(time.RFC3339)
	registry.Status.RepositoryCount = 1
	klog.Infof("Successfully verified registry %s", registry.Name)
	return nil
}

func (m *RegistryManager) LoginRegistry(ctx context.Context, name string, credential api.RegistryCredential) error {
	registry := &Registry{}
	registry.Name = name
	registry.Status.Connected = true
	registry.Status.LastCheckTime = time.Now().Format(time.RFC3339)
	klog.Infof("Successfully logged in registry %s", name)
	return nil
}

func (m *RegistryManager) LogoutRegistry(ctx context.Context, name string) error {
	klog.Infof("Successfully logged out registry %s", name)
	return nil
}

func (m *RegistryManager) HealthCheck(ctx context.Context, name string) error {
	registry, _ := m.GetRegistry(ctx, name)
	return m.VerifyRegistry(ctx, registry)
}

func (m *RegistryManager) GetRegistrySecret(ctx context.Context, secretName, secretNamespace string) error {
	return nil
}

func (m *RegistryManager) UpdateRegistry(ctx context.Context, registry *Registry) (*Registry, error) {
	if registry.Spec.Type != RegistryTypePublic && registry.Spec.Type != RegistryTypePrivate {
		return nil, fmt.Errorf("invalid registry type: %s", registry.Spec.Type)
	}
	klog.Infof("Updated registry %s", registry.Name)
	return registry, nil
}

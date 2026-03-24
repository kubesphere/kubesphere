/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package image_registry

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	imagev1alpha1 "kubesphere.io/kubesphere/pkg/apis/image-registry/v1alpha1"
)

func TestRegistryManager_CreateRegistry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = imagev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	manager := NewRegistryManager(fakeClient, fakeClient)

	ctx := context.Background()
	registry := &Registry{
		ObjectMeta: metav1.ObjectMeta{Name: "test-registry"},
		Spec: RegistrySpec{
			Type:   RegistryTypePublic,
			Domain: "docker.io",
		},
	}

	createdRegistry, err := manager.CreateRegistry(ctx, registry)
	if err != nil {
		t.Fatalf("failed to create registry: %v", err)
	}

	if createdRegistry.Name != "test-registry" {
		t.Errorf("expected registry name 'test-registry', got '%s'", createdRegistry.Name)
	}

	if createdRegistry.Spec.Type != RegistryTypePublic {
		t.Errorf("expected registry type '%s', got '%s'", RegistryTypePublic, createdRegistry.Spec.Type)
	}
}

func TestRegistryManager_GetRegistry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = imagev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	manager := NewRegistryManager(fakeClient, fakeClient)

	ctx := context.Background()

	// First create a registry
	registry := &Registry{
		ObjectMeta: metav1.ObjectMeta{Name: "test-registry"},
		Spec: RegistrySpec{
			Type:   RegistryTypePublic,
			Domain: "docker.io",
		},
	}
	_, _ = manager.CreateRegistry(ctx, registry)

	// Now try to get it
	retrievedRegistry, err := manager.GetRegistry(ctx, "test-registry")
	if err != nil {
		t.Fatalf("failed to get registry: %v", err)
	}

	if retrievedRegistry.Name != "test-registry" {
		t.Errorf("expected registry name 'test-registry', got '%s'", retrievedRegistry.Name)
	}
}

func TestRegistryManager_ListRegistries(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = imagev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	manager := NewRegistryManager(fakeClient, fakeClient)

	ctx := context.Background()

	// Create multiple registries
	registries := []Registry{
		{ObjectMeta: metav1.ObjectMeta{Name: "registry-1"}, Spec: RegistrySpec{Type: RegistryTypePublic, Domain: "docker.io"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "registry-2"}, Spec: RegistrySpec{Type: RegistryTypePrivate, Domain: "harbor.example.com"}},
	}

	for _, reg := range registries {
		_, _ = manager.CreateRegistry(ctx, &reg)
	}

	// List all registries
	listedRegistries, err := manager.ListRegistries(ctx)
	if err != nil {
		t.Fatalf("failed to list registries: %v", err)
	}

	if len(listedRegistries) != 2 {
		t.Errorf("expected 2 registries, got %d", len(listedRegistries))
	}
}

func TestRegistryManager_DeleteRegistry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = imagev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	manager := NewRegistryManager(fakeClient, fakeClient)

	ctx := context.Background()

	// Create a registry
	registry := &Registry{
		ObjectMeta: metav1.ObjectMeta{Name: "test-registry"},
		Spec: RegistrySpec{
			Type:   RegistryTypePublic,
			Domain: "docker.io",
		},
	}
	_, _ = manager.CreateRegistry(ctx, registry)

	// Delete the registry
	err := manager.DeleteRegistry(ctx, "test-registry")
	if err != nil {
		t.Fatalf("failed to delete registry: %v", err)
	}

	// Verify it's gone
	_, err = manager.GetRegistry(ctx, "test-registry")
	if err == nil {
		t.Error("expected error when getting deleted registry, got nil")
	}
}

func TestRegistryManager_UpdateRegistry(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = imagev1alpha1.AddToScheme(scheme)

	fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	manager := NewRegistryManager(fakeClient, fakeClient)

	ctx := context.Background()

	// Create a registry
	registry := &Registry{
		ObjectMeta: metav1.ObjectMeta{Name: "test-registry"},
		Spec: RegistrySpec{
			Type:   RegistryTypePublic,
			Domain: "docker.io",
		},
	}
	_, _ = manager.CreateRegistry(ctx, registry)

	// Update the registry
	updatedRegistry := &Registry{
		ObjectMeta: metav1.ObjectMeta{Name: "test-registry"},
		Spec: RegistrySpec{
			Type:     RegistryTypePrivate,
			Domain:   "harbor.example.com",
			Insecure: true,
		},
	}

	result, err := manager.UpdateRegistry(ctx, updatedRegistry)
	if err != nil {
		t.Fatalf("failed to update registry: %v", err)
	}

	if result.Spec.Type != RegistryTypePrivate {
		t.Errorf("expected registry type '%s', got '%s'", RegistryTypePrivate, result.Spec.Type)
	}

	if !result.Spec.Insecure {
		t.Error("expected registry to be insecure")
	}
}
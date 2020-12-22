// Copyright 2020 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package v1beta1 contains API Schema definitions for the app v1beta1 API group
// +kubebuilder:object:generate=true
// +groupName=app.k8s.io
package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "app.k8s.io", Version: "v1beta1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

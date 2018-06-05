/*
Copyright 2017 The Kubernetes Authors.

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

package customresource

import (
	"context"

	"github.com/go-openapi/validate"

	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	apiserverstorage "k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	utilfeature "k8s.io/apiserver/pkg/util/feature"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsfeatures "k8s.io/apiextensions-apiserver/pkg/features"
)

// customResourceStrategy implements behavior for CustomResources.
type customResourceStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator

	namespaceScoped bool
	validator       customResourceValidator
	status          *apiextensions.CustomResourceSubresourceStatus
	scale           *apiextensions.CustomResourceSubresourceScale
}

func NewStrategy(typer runtime.ObjectTyper, namespaceScoped bool, kind schema.GroupVersionKind, schemaValidator, statusSchemaValidator *validate.SchemaValidator, status *apiextensions.CustomResourceSubresourceStatus, scale *apiextensions.CustomResourceSubresourceScale) customResourceStrategy {
	return customResourceStrategy{
		ObjectTyper:     typer,
		NameGenerator:   names.SimpleNameGenerator,
		namespaceScoped: namespaceScoped,
		status:          status,
		scale:           scale,
		validator: customResourceValidator{
			namespaceScoped:       namespaceScoped,
			kind:                  kind,
			schemaValidator:       schemaValidator,
			statusSchemaValidator: statusSchemaValidator,
		},
	}
}

func (a customResourceStrategy) NamespaceScoped() bool {
	return a.namespaceScoped
}

// PrepareForCreate clears the status of a CustomResource before creation.
func (a customResourceStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
	if utilfeature.DefaultFeatureGate.Enabled(apiextensionsfeatures.CustomResourceSubresources) && a.status != nil {
		customResourceObject := obj.(*unstructured.Unstructured)
		customResource := customResourceObject.UnstructuredContent()

		// create cannot set status
		if _, ok := customResource["status"]; ok {
			delete(customResource, "status")
		}
	}

	accessor, _ := meta.Accessor(obj)
	accessor.SetGeneration(1)
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (a customResourceStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	if !utilfeature.DefaultFeatureGate.Enabled(apiextensionsfeatures.CustomResourceSubresources) || a.status == nil {
		return
	}

	newCustomResourceObject := obj.(*unstructured.Unstructured)
	oldCustomResourceObject := old.(*unstructured.Unstructured)

	newCustomResource := newCustomResourceObject.UnstructuredContent()
	oldCustomResource := oldCustomResourceObject.UnstructuredContent()

	// update is not allowed to set status
	_, ok1 := newCustomResource["status"]
	_, ok2 := oldCustomResource["status"]
	switch {
	case ok2:
		newCustomResource["status"] = oldCustomResource["status"]
	case ok1:
		delete(newCustomResource, "status")
	}

	// Any changes to the spec increment the generation number, any changes to the
	// status should reflect the generation number of the corresponding object. We push
	// the burden of managing the status onto the clients because we can't (in general)
	// know here what version of spec the writer of the status has seen. It may seem like
	// we can at first -- since obj contains spec -- but in the future we will probably make
	// status its own object, and even if we don't, writes may be the result of a
	// read-update-write loop, so the contents of spec may not actually be the spec that
	// the CustomResource has *seen*.
	newSpec, ok1 := newCustomResource["spec"]
	oldSpec, ok2 := oldCustomResource["spec"]

	// spec is changed, created or deleted
	if (ok1 && ok2 && !apiequality.Semantic.DeepEqual(oldSpec, newSpec)) || (ok1 && !ok2) || (!ok1 && ok2) {
		oldAccessor, _ := meta.Accessor(oldCustomResourceObject)
		newAccessor, _ := meta.Accessor(newCustomResourceObject)
		newAccessor.SetGeneration(oldAccessor.GetGeneration() + 1)
	}
}

// Validate validates a new CustomResource.
func (a customResourceStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	return a.validator.Validate(ctx, obj, a.scale)
}

// Canonicalize normalizes the object after validation.
func (customResourceStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for CustomResources; this means a POST is
// needed to create one.
func (customResourceStrategy) AllowCreateOnUpdate() bool {
	return false
}

// AllowUnconditionalUpdate is the default update policy for CustomResource objects.
func (customResourceStrategy) AllowUnconditionalUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user updating status.
func (a customResourceStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return a.validator.ValidateUpdate(ctx, obj, old, a.scale)
}

// GetAttrs returns labels and fields of a given object for filtering purposes.
func (a customResourceStrategy) GetAttrs(obj runtime.Object) (labels.Set, fields.Set, bool, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, nil, false, err
	}
	return labels.Set(accessor.GetLabels()), objectMetaFieldsSet(accessor, a.namespaceScoped), accessor.GetInitializers() != nil, nil
}

// objectMetaFieldsSet returns a fields that represent the ObjectMeta.
func objectMetaFieldsSet(objectMeta metav1.Object, namespaceScoped bool) fields.Set {
	if namespaceScoped {
		return fields.Set{
			"metadata.name":      objectMeta.GetName(),
			"metadata.namespace": objectMeta.GetNamespace(),
		}
	}
	return fields.Set{
		"metadata.name": objectMeta.GetName(),
	}
}

// MatchCustomResourceDefinitionStorage is the filter used by the generic etcd backend to route
// watch events from etcd to clients of the apiserver only interested in specific
// labels/fields.
func (a customResourceStrategy) MatchCustomResourceDefinitionStorage(label labels.Selector, field fields.Selector) apiserverstorage.SelectionPredicate {
	return apiserverstorage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: a.GetAttrs,
	}
}

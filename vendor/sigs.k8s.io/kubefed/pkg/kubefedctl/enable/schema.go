/*
Copyright 2018 The Kubernetes Authors.

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

package enable

import (
	"fmt"

	"github.com/pkg/errors"

	apiextv1b1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextv1b1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/kube-openapi/pkg/util/proto"
	"k8s.io/kubectl/pkg/util/openapi"
)

type schemaAccessor interface {
	templateSchema() map[string]apiextv1b1.JSONSchemaProps
}

func newSchemaAccessor(config *rest.Config, apiResource metav1.APIResource) (schemaAccessor, error) {
	// Assume the resource may be a CRD, and fall back to OpenAPI if that is not the case.
	crdAccessor, err := newCRDSchemaAccessor(config, apiResource)
	if err != nil {
		return nil, err
	}
	if crdAccessor != nil {
		return crdAccessor, nil
	}
	return newOpenAPISchemaAccessor(config, apiResource)
}

type crdSchemaAccessor struct {
	validation *apiextv1b1.CustomResourceValidation
}

func newCRDSchemaAccessor(config *rest.Config, apiResource metav1.APIResource) (schemaAccessor, error) {
	// CRDs must have a group
	if len(apiResource.Group) == 0 {
		return nil, nil
	}
	// Check whether the target resource is a crd
	crdClient, err := apiextv1b1client.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create crd clientset")
	}
	crdName := fmt.Sprintf("%s.%s", apiResource.Name, apiResource.Group)
	crd, err := crdClient.CustomResourceDefinitions().Get(crdName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "Error attempting retrieval of crd %q", crdName)
	}
	return &crdSchemaAccessor{validation: crd.Spec.Validation}, nil
}

func (a *crdSchemaAccessor) templateSchema() map[string]apiextv1b1.JSONSchemaProps {
	if a.validation != nil && a.validation.OpenAPIV3Schema != nil {
		return a.validation.OpenAPIV3Schema.Properties
	}
	return nil
}

type openAPISchemaAccessor struct {
	targetResource proto.Schema
}

func newOpenAPISchemaAccessor(config *rest.Config, apiResource metav1.APIResource) (schemaAccessor, error) {
	client, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "Error creating discovery client")
	}
	resources, err := openapi.NewOpenAPIGetter(client).Get()
	if err != nil {
		return nil, errors.Wrap(err, "Error loading openapi schema")
	}
	gvk := schema.GroupVersionKind{
		Group:   apiResource.Group,
		Version: apiResource.Version,
		Kind:    apiResource.Kind,
	}
	targetResource := resources.LookupResource(gvk)
	if targetResource == nil {
		return nil, errors.Errorf("Unable to find openapi schema for %q", gvk)
	}
	return &openAPISchemaAccessor{
		targetResource: targetResource,
	}, nil
}

func (a *openAPISchemaAccessor) templateSchema() map[string]apiextv1b1.JSONSchemaProps {
	var templateSchema *apiextv1b1.JSONSchemaProps
	visitor := &jsonSchemaVistor{
		collect: func(schema apiextv1b1.JSONSchemaProps) {
			templateSchema = &schema
		},
	}
	a.targetResource.Accept(visitor)

	return templateSchema.Properties
}

// jsonSchemaVistor converts proto.Schema resources into json schema.
// A local visitor (and associated callback) is intended to be created
// whenever a function needs to recurse.
//
// TODO(marun) Generate more extensive schema if/when openapi schema
// provides more detail as per https://github.com/ant31/crd-validation
type jsonSchemaVistor struct {
	collect func(schema apiextv1b1.JSONSchemaProps)
}

func (v *jsonSchemaVistor) VisitArray(a *proto.Array) {
	arraySchema := apiextv1b1.JSONSchemaProps{
		Type:  "array",
		Items: &apiextv1b1.JSONSchemaPropsOrArray{},
	}
	localVisitor := &jsonSchemaVistor{
		collect: func(schema apiextv1b1.JSONSchemaProps) {
			arraySchema.Items.Schema = &schema
		},
	}
	a.SubType.Accept(localVisitor)
	v.collect(arraySchema)
}

func (v *jsonSchemaVistor) VisitMap(m *proto.Map) {
	mapSchema := apiextv1b1.JSONSchemaProps{
		Type: "object",
		AdditionalProperties: &apiextv1b1.JSONSchemaPropsOrBool{
			Allows: true,
		},
	}
	localVisitor := &jsonSchemaVistor{
		collect: func(schema apiextv1b1.JSONSchemaProps) {
			mapSchema.AdditionalProperties.Schema = &schema
		},
	}
	m.SubType.Accept(localVisitor)
	v.collect(mapSchema)
}

func (v *jsonSchemaVistor) VisitPrimitive(p *proto.Primitive) {
	schema := schemaForPrimitive(p)
	v.collect(schema)
}

func (v *jsonSchemaVistor) VisitKind(k *proto.Kind) {
	kindSchema := apiextv1b1.JSONSchemaProps{
		Type:       "object",
		Properties: make(map[string]apiextv1b1.JSONSchemaProps),
		Required:   k.RequiredFields,
	}
	for key, fieldSchema := range k.Fields {
		// Status cannot be defined for a template
		if key == "status" {
			continue
		}
		localVisitor := &jsonSchemaVistor{
			collect: func(schema apiextv1b1.JSONSchemaProps) {
				kindSchema.Properties[key] = schema
			},
		}
		fieldSchema.Accept(localVisitor)
	}
	v.collect(kindSchema)
}

func (v *jsonSchemaVistor) VisitReference(r proto.Reference) {
	// Short-circuit the recursive definition of JSONSchemaProps (used for CRD validation)
	//
	// TODO(marun) Implement proper support for recursive schema
	if r.Reference() == "io.k8s.apiextensions-apiserver.pkg.apis.apiextensions.v1beta1.JSONSchemaProps" {
		v.collect(apiextv1b1.JSONSchemaProps{Type: "object"})
		return
	}

	r.SubSchema().Accept(v)
}

func schemaForPrimitive(p *proto.Primitive) apiextv1b1.JSONSchemaProps {
	schema := apiextv1b1.JSONSchemaProps{}

	if p.Format == "int-or-string" {
		schema.AnyOf = []apiextv1b1.JSONSchemaProps{
			{
				Type:   "integer",
				Format: "int32",
			},
			{
				Type: "string",
			},
		}
		return schema
	}

	if len(p.Format) > 0 {
		schema.Format = p.Format
	}
	schema.Type = p.Type
	return schema
}

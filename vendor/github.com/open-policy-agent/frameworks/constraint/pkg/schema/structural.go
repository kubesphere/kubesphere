package schema

import (
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/schema"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

var constraintTemplateCRD *apiextensionsv1.CustomResourceDefinition

func init() {
	// Ingest the constraint template CRD for use in defaulting functions
	constraintTemplateCRD = &apiextensionsv1.CustomResourceDefinition{}
	if err := yaml.Unmarshal([]byte(constraintTemplateCRDYaml), constraintTemplateCRD); err != nil {
		panic(fmt.Errorf("%w: failed to unmarshal yaml into constraintTemplateCRD", err))
	}
}

func CRDSchema(sch *runtime.Scheme, version string) (*schema.Structural, error) {
	// Fill version map with Structural types derived from ConstraintTemplate versions
	for _, crdVersion := range constraintTemplateCRD.Spec.Versions {
		if crdVersion.Name != version {
			continue
		}

		versionlessSchema := &apiextensions.JSONSchemaProps{}
		err := sch.Convert(crdVersion.Schema.OpenAPIV3Schema, versionlessSchema, nil)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to convert JSONSchemaProps for ConstraintTemplate version %v", err, crdVersion.Name)
		}

		structural, err := schema.NewStructural(versionlessSchema)
		if err != nil {
			return nil, fmt.Errorf("%w: failed to create Structural for ConstraintTemplate version %v", err, crdVersion.Name)
		}

		return structural, nil
	}

	return nil, fmt.Errorf("No CRD version '%q'", version)
}

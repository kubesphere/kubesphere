/*

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

package v1beta1

import (
	coreTemplates "github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	"github.com/open-policy-agent/frameworks/constraint/pkg/schema"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/conversion"
)

func Convert_v1beta1_Validation_To_templates_Validation(in *Validation, out *coreTemplates.Validation, s conversion.Scope) error { // nolint:revive // Required exact function name.
	inSchema := in.OpenAPIV3Schema

	// legacySchema should allow for users to provide arbitrary parameters, regardless of whether the user specified them
	if in.LegacySchema != nil && *in.LegacySchema && inSchema == nil {
		inSchema = &apiextensionsv1.JSONSchemaProps{}
	}

	if inSchema != nil {
		inSchemaCopy := inSchema.DeepCopy()

		if in.LegacySchema != nil && *in.LegacySchema {
			if err := schema.AddPreserveUnknownFields(inSchemaCopy); err != nil {
				return err
			}
		}

		out.OpenAPIV3Schema = new(apiextensions.JSONSchemaProps)
		if err := apiextensionsv1.Convert_v1_JSONSchemaProps_To_apiextensions_JSONSchemaProps(inSchemaCopy, out.OpenAPIV3Schema, s); err != nil {
			return err
		}
	} else {
		out.OpenAPIV3Schema = nil
	}

	// As LegacySchema is a pointer, we have to explicitly copy the value.  Doing a simple copy of
	// out.LegacySchema = in.LegacySchema yields a duplicate pointer to the same value.  This links
	// the value of LegacySchema in the out object to that of the in object, potentially creating
	// a bug where both change when only one is meant to.
	if in.LegacySchema == nil {
		out.LegacySchema = nil
	} else {
		inVal := *in.LegacySchema
		out.LegacySchema = &inVal
	}

	return nil
}

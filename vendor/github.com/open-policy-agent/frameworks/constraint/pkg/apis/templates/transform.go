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

package templates

import (
	"fmt"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

// AddPreserveUnknownFields recurses through an *apiextensionsv1beta1.JSONSchemaProps
// data structure, adding `x-kubernetes-preserve-unknown-fields: true` at every level
// that type is equal to "object", "array", or is undefined.
func AddPreserveUnknownFields(sch *apiextensionsv1beta1.JSONSchemaProps) error {
	switch sch.Type {
	// An object can have values not described in the schema.  A blank Type could be anything,
	// including an object, so we add x-kubernetes-preserve-unknown-fields: true to both.
	case "", "object":
		tr := true
		sch.XPreserveUnknownFields = &tr
	case "array":
		// If the type is array, the schema of the array's items must be structural.  If the schema
		// is undefined, we must add a blank one with x-kubernetes-preserve-unknown-fields to meet
		// this structural item schema requirement.
		if sch.Items == nil || (sch.Items.Schema == nil && sch.Items.JSONSchemas == nil) {
			tr := true
			sch.Items = &apiextensionsv1beta1.JSONSchemaPropsOrArray{
				Schema: &apiextensionsv1beta1.JSONSchemaProps{
					XPreserveUnknownFields: &tr,
				},
			}
		}
	}

	if sch.Properties != nil {
		for k, v := range sch.Properties {
			if err := AddPreserveUnknownFields(&v); err != nil {
				return err
			}

			// As v is not a pointer, we need to set the contents of v back into the original data structure
			sch.Properties[k] = v
		}
	}

	if sch.Items != nil {
		if sch.Items.Schema != nil {
			if err := AddPreserveUnknownFields(sch.Items.Schema); err != nil {
				return err
			}
		}

		if sch.Items.JSONSchemas != nil {
			return fmt.Errorf("non-nil JSONSchemas encountered, multiple schemas are not supported")
		}
	}

	if sch.AdditionalProperties != nil && sch.AdditionalProperties.Schema != nil {
		if err := AddPreserveUnknownFields(sch.AdditionalProperties.Schema); err != nil {
			return err
		}
	}

	return nil
}

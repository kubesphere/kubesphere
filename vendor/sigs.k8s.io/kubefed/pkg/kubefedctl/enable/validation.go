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
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	"sigs.k8s.io/kubefed/pkg/controller/util"
)

func federatedTypeValidationSchema(templateSchema map[string]v1beta1.JSONSchemaProps) *v1beta1.CustomResourceValidation {
	schema := ValidationSchema(v1beta1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]v1beta1.JSONSchemaProps{
			"placement": {
				Type: "object",
				Properties: map[string]v1beta1.JSONSchemaProps{
					// References to one or more clusters allow a
					// scheduling mechanism to explicitly indicate
					// placement. If one or more clusters is provided,
					// the clusterSelector field will be ignored.
					"clusters": {
						Type: "array",
						Items: &v1beta1.JSONSchemaPropsOrArray{
							Schema: &v1beta1.JSONSchemaProps{
								Type: "object",
								Properties: map[string]v1beta1.JSONSchemaProps{
									"name": {
										Type: "string",
									},
								},
								Required: []string{
									"name",
								},
							},
						},
					},
					"clusterSelector": {
						Type: "object",
						Properties: map[string]v1beta1.JSONSchemaProps{
							"matchExpressions": {
								Type: "array",
								Items: &v1beta1.JSONSchemaPropsOrArray{
									Schema: &v1beta1.JSONSchemaProps{
										Type: "object",
										Properties: map[string]v1beta1.JSONSchemaProps{
											"key": {
												Type: "string",
											},
											"operator": {
												Type: "string",
											},
											"values": {
												Type: "array",
												Items: &v1beta1.JSONSchemaPropsOrArray{
													Schema: &v1beta1.JSONSchemaProps{
														Type: "string",
													},
												},
											},
										},
										Required: []string{
											"key",
											"operator",
										},
									},
								},
							},
							"matchLabels": {
								Type: "object",
								AdditionalProperties: &v1beta1.JSONSchemaPropsOrBool{
									Schema: &v1beta1.JSONSchemaProps{
										Type: "string",
									},
								},
							},
						},
					},
				},
			},
			"overrides": {
				Type: "array",
				Items: &v1beta1.JSONSchemaPropsOrArray{
					Schema: &v1beta1.JSONSchemaProps{
						Type: "object",
						Properties: map[string]v1beta1.JSONSchemaProps{
							"clusterName": {
								Type: "string",
							},
							"clusterOverrides": {
								Type: "array",
								Items: &v1beta1.JSONSchemaPropsOrArray{
									Schema: &v1beta1.JSONSchemaProps{
										Type: "object",
										Properties: map[string]v1beta1.JSONSchemaProps{
											"op": {
												Type:    "string",
												Pattern: "^(add|remove|replace)?$",
											},
											"path": {
												Type: "string",
											},
											"value": {
												// Supporting the override of an arbitrary field
												// precludes up-front validation.  Errors in
												// the definition of override values will need to
												// be caught during propagation.
												AnyOf: []v1beta1.JSONSchemaProps{
													{
														Type: "string",
													},
													{
														Type: "integer",
													},
													{
														Type: "boolean",
													},
													{
														Type: "object",
													},
													{
														Type: "array",
													},
												},
											},
										},
										Required: []string{
											"path",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	})
	if templateSchema != nil {
		specProperties := schema.OpenAPIV3Schema.Properties["spec"].Properties
		specProperties["template"] = v1beta1.JSONSchemaProps{
			Type: "object",
		}
		// Add retainReplicas field to types that exposes a replicas
		// field that could be targeted by HPA.
		if templateSpec, ok := templateSchema["spec"]; ok {
			// TODO: find a simpler way to detect that a resource is scalable than having to compute the entire schema.
			if replicasField, ok := templateSpec.Properties["replicas"]; ok {
				if replicasField.Type == "integer" && replicasField.Format == "int32" {
					specProperties[util.RetainReplicasField] = v1beta1.JSONSchemaProps{
						Type: "boolean",
					}
				}
			}
		}

	}
	return schema
}

func ValidationSchema(specProps v1beta1.JSONSchemaProps) *v1beta1.CustomResourceValidation {
	return &v1beta1.CustomResourceValidation{
		OpenAPIV3Schema: &v1beta1.JSONSchemaProps{
			Properties: map[string]v1beta1.JSONSchemaProps{
				"apiVersion": {
					Type: "string",
				},
				"kind": {
					Type: "string",
				},
				// TODO(marun) Add a comprehensive schema for metadata
				"metadata": {
					Type: "object",
				},
				"spec": specProps,
				"status": {
					Type: "object",
					Properties: map[string]v1beta1.JSONSchemaProps{
						"conditions": {
							Type: "array",
							Items: &v1beta1.JSONSchemaPropsOrArray{
								Schema: &v1beta1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"type": {
											Type: "string",
										},
										"status": {
											Type: "string",
										},
										"reason": {
											Type: "string",
										},
										"lastUpdateTime": {
											Format: "date-time",
											Type:   "string",
										},
										"lastTransitionTime": {
											Format: "date-time",
											Type:   "string",
										},
									},
									Required: []string{
										"type",
										"status",
									},
								},
							},
						},
						"clusters": {
							Type: "array",
							Items: &v1beta1.JSONSchemaPropsOrArray{
								Schema: &v1beta1.JSONSchemaProps{
									Type: "object",
									Properties: map[string]v1beta1.JSONSchemaProps{
										"name": {
											Type: "string",
										},
										"status": {
											Type: "string",
										},
									},
									Required: []string{
										"name",
									},
								},
							},
						},
						"observedGeneration": {
							Format: "int64",
							Type:   "integer",
						},
					},
				},
			},
			// Require a spec (even if empty) as an aid to users
			// manually creating federated configmaps or
			// secrets. These target types do not include a spec,
			// and the absence of the spec in a federated
			// equivalent could indicate a malformed resource.
			Required: []string{
				"spec",
			},
		},
	}
}

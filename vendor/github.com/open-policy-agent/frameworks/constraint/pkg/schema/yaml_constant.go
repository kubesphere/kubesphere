package schema

// This file is generated from deploy/crds.yaml via "make constraint-template-string-constant"
// DO NOT MODIFY THIS FILE DIRECTLY!

const constraintTemplateCRDYaml = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.5.0
  name: constrainttemplates.templates.gatekeeper.sh
spec:
  group: templates.gatekeeper.sh
  names:
    kind: ConstraintTemplate
    listKind: ConstraintTemplateList
    plural: constrainttemplates
    singular: constrainttemplate
  preserveUnknownFields: false
  scope: Cluster
  versions:
  - name: v1
    schema:
      openAPIV3Schema:
        description: ConstraintTemplate is the Schema for the constrainttemplates API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ConstraintTemplateSpec defines the desired state of ConstraintTemplate.
            properties:
              crd:
                properties:
                  spec:
                    properties:
                      names:
                        properties:
                          kind:
                            type: string
                          shortNames:
                            items:
                              type: string
                            type: array
                        type: object
                      validation:
                        default:
                          legacySchema: false
                        properties:
                          legacySchema:
                            default: false
                            type: boolean
                          openAPIV3Schema:
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                        type: object
                    type: object
                type: object
              targets:
                items:
                  properties:
                    libs:
                      items:
                        type: string
                      type: array
                    rego:
                      type: string
                    target:
                      type: string
                  type: object
                type: array
            type: object
          status:
            description: ConstraintTemplateStatus defines the observed state of ConstraintTemplate.
            properties:
              byPod:
                items:
                  description: ByPodStatus defines the observed state of ConstraintTemplate as seen by an individual controller
                  properties:
                    errors:
                      items:
                        description: CreateCRDError represents a single error caught during parsing, compiling, etc.
                        properties:
                          code:
                            type: string
                          location:
                            type: string
                          message:
                            type: string
                        required:
                        - code
                        - message
                        type: object
                      type: array
                    id:
                      description: a unique identifier for the pod that wrote the status
                      type: string
                    observedGeneration:
                      format: int64
                      type: integer
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                type: array
              created:
                type: boolean
            type: object
        type: object
    served: true
    storage: true
    subresources:
      status: {}
  - name: v1alpha1
    schema:
      openAPIV3Schema:
        description: ConstraintTemplate is the Schema for the constrainttemplates API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ConstraintTemplateSpec defines the desired state of ConstraintTemplate.
            properties:
              crd:
                properties:
                  spec:
                    properties:
                      names:
                        properties:
                          kind:
                            type: string
                          shortNames:
                            items:
                              type: string
                            type: array
                        type: object
                      validation:
                        default:
                          legacySchema: true
                        properties:
                          legacySchema:
                            default: true
                            type: boolean
                          openAPIV3Schema:
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                        type: object
                    type: object
                type: object
              targets:
                items:
                  properties:
                    libs:
                      items:
                        type: string
                      type: array
                    rego:
                      type: string
                    target:
                      type: string
                  type: object
                type: array
            type: object
          status:
            description: ConstraintTemplateStatus defines the observed state of ConstraintTemplate.
            properties:
              byPod:
                items:
                  description: ByPodStatus defines the observed state of ConstraintTemplate as seen by an individual controller
                  properties:
                    errors:
                      items:
                        description: CreateCRDError represents a single error caught during parsing, compiling, etc.
                        properties:
                          code:
                            type: string
                          location:
                            type: string
                          message:
                            type: string
                        required:
                        - code
                        - message
                        type: object
                      type: array
                    id:
                      description: a unique identifier for the pod that wrote the status
                      type: string
                    observedGeneration:
                      format: int64
                      type: integer
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                type: array
              created:
                type: boolean
            type: object
        type: object
    served: true
    storage: false
    subresources:
      status: {}
  - name: v1beta1
    schema:
      openAPIV3Schema:
        description: ConstraintTemplate is the Schema for the constrainttemplates API
        properties:
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          kind:
            description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          spec:
            description: ConstraintTemplateSpec defines the desired state of ConstraintTemplate.
            properties:
              crd:
                properties:
                  spec:
                    properties:
                      names:
                        properties:
                          kind:
                            type: string
                          shortNames:
                            items:
                              type: string
                            type: array
                        type: object
                      validation:
                        default:
                          legacySchema: true
                        properties:
                          legacySchema:
                            default: true
                            type: boolean
                          openAPIV3Schema:
                            type: object
                            x-kubernetes-preserve-unknown-fields: true
                        type: object
                    type: object
                type: object
              targets:
                items:
                  properties:
                    libs:
                      items:
                        type: string
                      type: array
                    rego:
                      type: string
                    target:
                      type: string
                  type: object
                type: array
            type: object
          status:
            description: ConstraintTemplateStatus defines the observed state of ConstraintTemplate.
            properties:
              byPod:
                items:
                  description: ByPodStatus defines the observed state of ConstraintTemplate as seen by an individual controller
                  properties:
                    errors:
                      items:
                        description: CreateCRDError represents a single error caught during parsing, compiling, etc.
                        properties:
                          code:
                            type: string
                          location:
                            type: string
                          message:
                            type: string
                        required:
                        - code
                        - message
                        type: object
                      type: array
                    id:
                      description: a unique identifier for the pod that wrote the status
                      type: string
                    observedGeneration:
                      format: int64
                      type: integer
                  type: object
                  x-kubernetes-preserve-unknown-fields: true
                type: array
              created:
                type: boolean
            type: object
        type: object
    served: true
    storage: false
    subresources:
      status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`

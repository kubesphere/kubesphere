package provider

import (
	"context"
	"encoding/json"
	"github.com/open-policy-agent/frameworks/constraint/pkg/client"
	"github.com/open-policy-agent/frameworks/constraint/pkg/core/templates"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"kubesphere.io/api/admission/v1alpha1"
	"strings"
)

const (
	GateKeeperProviderName = "gatekeeper"
)

var (
	ConstraintsGroup   = "constraints.gatekeeper.sh"
	ConstraintsVersion = "v1alpha1"
)

type GateKeeperProvider struct {
	*client.Client
}

func NewGateKeeperProvider(client *client.Client) *GateKeeperProvider {
	return &GateKeeperProvider{client}
}

func (p *GateKeeperProvider) AddPolicy(ctx context.Context, policy *v1alpha1.Policy) error {
	template, err := Template(policy)
	if err != nil {
		return err
	}
	_, err = p.AddTemplate(ctx, template)
	if err != nil {
		return err
	}
	return nil
}

func (p *GateKeeperProvider) AddRule(ctx context.Context, rule *v1alpha1.Rule) error {
	constraint, err := Constraint(rule)
	if err != nil {
		return err
	}
	_, err = p.AddConstraint(ctx, constraint)
	if err != nil {
		return err
	}
	return nil
}

func (p *GateKeeperProvider) RemovePolicy(ctx context.Context, policy *v1alpha1.Policy) error {
	template, err := Template(policy)
	if err != nil {
		return err
	}
	_, err = p.RemoveTemplate(ctx, template)
	if err != nil {
		return err
	}
	return nil
}

func (p *GateKeeperProvider) RemoveRule(ctx context.Context, rule *v1alpha1.Rule) error {
	constraint, err := Constraint(rule)
	if err != nil {
		return err
	}
	_, err = p.RemoveConstraint(ctx, constraint)
	if err != nil {
		return err
	}
	return nil
}

func Template(policy *v1alpha1.Policy) (*templates.ConstraintTemplate, error) {
	spec := policy.Spec.Content.Spec
	targets := policy.Spec.Content.Targets

	var templateTargets []templates.Target
	for _, target := range targets {
		templateTargets = append(templateTargets, templates.Target{
			Target: target.Target,
			Rego:   target.Expression,
			Libs:   target.Import,
		})
	}

	template := &templates.ConstraintTemplate{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: strings.ToLower(policy.Name),
		},
		Spec: templates.ConstraintTemplateSpec{
			CRD: templates.CRD{
				Spec: templates.CRDSpec{
					Names: templates.Names{
						Kind:       spec.Names.Name,
						ShortNames: spec.Names.ShortNames,
					},
					Validation: &templates.Validation{
						OpenAPIV3Schema: spec.Parameters.Validation.OpenAPIV3Schema,
						LegacySchema:    &spec.Parameters.Validation.LegacySchema,
					},
				},
			},
			Targets: templateTargets,
		},
		Status: templates.ConstraintTemplateStatus{},
	}

	return template, nil
}

func Constraint(rule *v1alpha1.Rule) (*unstructured.Unstructured, error) {
	c := &unstructured.Unstructured{}
	c.SetGroupVersionKind(ConstraintGvk(rule.Name))
	paramMap := map[string]interface{}{}
	err := json.Unmarshal(rule.Spec.Parameters.Raw, &paramMap)
	if err := unstructured.SetNestedMap(c.Object, paramMap, "spec", "parameters"); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return c, nil
}

func ConstraintGvk(kind string) schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   ConstraintsGroup,
		Version: ConstraintsVersion,
		Kind:    kind,
	}
}

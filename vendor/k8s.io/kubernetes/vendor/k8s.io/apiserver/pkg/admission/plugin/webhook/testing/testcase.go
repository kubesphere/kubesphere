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

package testing

import (
	"net/url"

	registrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/plugin/webhook/testcerts"
	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
)

var matchEverythingRules = []registrationv1beta1.RuleWithOperations{{
	Operations: []registrationv1beta1.OperationType{registrationv1beta1.OperationAll},
	Rule: registrationv1beta1.Rule{
		APIGroups:   []string{"*"},
		APIVersions: []string{"*"},
		Resources:   []string{"*/*"},
	},
}}

// NewFakeDataSource returns a mock client and informer returning the given webhooks.
func NewFakeDataSource(name string, webhooks []registrationv1beta1.Webhook, mutating bool, stopCh <-chan struct{}) (clientset kubernetes.Interface, factory informers.SharedInformerFactory) {
	var objs = []runtime.Object{
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					"runlevel": "0",
				},
			},
		},
	}
	if mutating {
		objs = append(objs, &registrationv1beta1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-webhooks",
			},
			Webhooks: webhooks,
		})
	} else {
		objs = append(objs, &registrationv1beta1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-webhooks",
			},
			Webhooks: webhooks,
		})
	}

	client := fakeclientset.NewSimpleClientset(objs...)
	informerFactory := informers.NewSharedInformerFactory(client, 0)

	return client, informerFactory
}

// NewAttribute returns static admission Attributes for testing.
func NewAttribute(namespace string) admission.Attributes {
	// Set up a test object for the call
	kind := corev1.SchemeGroupVersion.WithKind("Pod")
	name := "my-pod"
	object := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"pod.name": name,
			},
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
	}
	oldObject := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
	}
	operation := admission.Update
	resource := corev1.Resource("pods").WithVersion("v1")
	subResource := ""
	userInfo := user.DefaultInfo{
		Name: "webhook-test",
		UID:  "webhook-test",
	}

	return admission.NewAttributesRecord(&object, &oldObject, kind, namespace, name, resource, subResource, operation, &userInfo)
}

type urlConfigGenerator struct {
	baseURL *url.URL
}

func (c urlConfigGenerator) ccfgURL(urlPath string) registrationv1beta1.WebhookClientConfig {
	u2 := *c.baseURL
	u2.Path = urlPath
	urlString := u2.String()
	return registrationv1beta1.WebhookClientConfig{
		URL:      &urlString,
		CABundle: testcerts.CACert,
	}
}

// Test is a webhook test case.
type Test struct {
	Name          string
	Webhooks      []registrationv1beta1.Webhook
	Path          string
	ExpectAllow   bool
	ErrorContains string
}

// NewTestCases returns test cases with a given base url.
func NewTestCases(url *url.URL) []Test {
	policyFail := registrationv1beta1.Fail
	policyIgnore := registrationv1beta1.Ignore
	ccfgURL := urlConfigGenerator{url}.ccfgURL

	return []Test{
		{
			Name: "no match",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:         "nomatch",
				ClientConfig: ccfgSVC("disallow"),
				Rules: []registrationv1beta1.RuleWithOperations{{
					Operations: []registrationv1beta1.OperationType{registrationv1beta1.Create},
				}},
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ExpectAllow: true,
		},
		{
			Name: "match & allow",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "allow",
				ClientConfig:      ccfgSVC("allow"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ExpectAllow: true,
		},
		{
			Name: "match & disallow",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "disallow",
				ClientConfig:      ccfgSVC("disallow"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ErrorContains: "without explanation",
		},
		{
			Name: "match & disallow ii",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "disallowReason",
				ClientConfig:      ccfgSVC("disallowReason"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},

			ErrorContains: "you shall not pass",
		},
		{
			Name: "match & disallow & but allowed because namespaceSelector exempt the ns",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:         "disallow",
				ClientConfig: ccfgSVC("disallow"),
				Rules:        newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "runlevel",
						Values:   []string{"1"},
						Operator: metav1.LabelSelectorOpIn,
					}},
				},
			}},

			ExpectAllow: true,
		},
		{
			Name: "match & disallow & but allowed because namespaceSelector exempt the ns ii",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:         "disallow",
				ClientConfig: ccfgSVC("disallow"),
				Rules:        newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{{
						Key:      "runlevel",
						Values:   []string{"0"},
						Operator: metav1.LabelSelectorOpNotIn,
					}},
				},
			}},
			ExpectAllow: true,
		},
		{
			Name: "match & fail (but allow because fail open)",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "internalErr A",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}, {
				Name:              "internalErr B",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}, {
				Name:              "internalErr C",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},

			ExpectAllow: true,
		},
		{
			Name: "match & fail (but disallow because fail close on nil FailurePolicy)",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "internalErr A",
				ClientConfig:      ccfgSVC("internalErr"),
				NamespaceSelector: &metav1.LabelSelector{},
				Rules:             matchEverythingRules,
			}, {
				Name:              "internalErr B",
				ClientConfig:      ccfgSVC("internalErr"),
				NamespaceSelector: &metav1.LabelSelector{},
				Rules:             matchEverythingRules,
			}, {
				Name:              "internalErr C",
				ClientConfig:      ccfgSVC("internalErr"),
				NamespaceSelector: &metav1.LabelSelector{},
				Rules:             matchEverythingRules,
			}},
			ExpectAllow: false,
		},
		{
			Name: "match & fail (but fail because fail closed)",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "internalErr A",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyFail,
			}, {
				Name:              "internalErr B",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyFail,
			}, {
				Name:              "internalErr C",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyFail,
			}},
			ExpectAllow: false,
		},
		{
			Name: "match & allow (url)",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "allow",
				ClientConfig:      ccfgURL("allow"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ExpectAllow: true,
		},
		{
			Name: "match & disallow (url)",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "disallow",
				ClientConfig:      ccfgURL("disallow"),
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ErrorContains: "without explanation",
		}, {
			Name: "absent response and fail open",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "nilResponse",
				ClientConfig:      ccfgURL("nilResponse"),
				FailurePolicy:     &policyIgnore,
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ExpectAllow: true,
		},
		{
			Name: "absent response and fail closed",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "nilResponse",
				ClientConfig:      ccfgURL("nilResponse"),
				FailurePolicy:     &policyFail,
				Rules:             matchEverythingRules,
				NamespaceSelector: &metav1.LabelSelector{},
			}},
			ErrorContains: "Webhook response was absent",
		},
		// No need to test everything with the url case, since only the
		// connection is different.
	}
}

// CachedTest is a test case for the client manager.
type CachedTest struct {
	Name            string
	Webhooks        []registrationv1beta1.Webhook
	ExpectAllow     bool
	ExpectCacheMiss bool
}

// NewCachedClientTestcases returns a set of client manager test cases.
func NewCachedClientTestcases(url *url.URL) []CachedTest {
	policyIgnore := registrationv1beta1.Ignore
	ccfgURL := urlConfigGenerator{url}.ccfgURL

	return []CachedTest{
		{
			Name: "uncached: service webhook, path 'allow'",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "cache1",
				ClientConfig:      ccfgSVC("allow"),
				Rules:             newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},
			ExpectAllow:     true,
			ExpectCacheMiss: true,
		},
		{
			Name: "uncached: service webhook, path 'internalErr'",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "cache2",
				ClientConfig:      ccfgSVC("internalErr"),
				Rules:             newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},
			ExpectAllow:     true,
			ExpectCacheMiss: true,
		},
		{
			Name: "cached: service webhook, path 'allow'",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "cache3",
				ClientConfig:      ccfgSVC("allow"),
				Rules:             newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},
			ExpectAllow:     true,
			ExpectCacheMiss: false,
		},
		{
			Name: "uncached: url webhook, path 'allow'",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "cache4",
				ClientConfig:      ccfgURL("allow"),
				Rules:             newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},
			ExpectAllow:     true,
			ExpectCacheMiss: true,
		},
		{
			Name: "cached: service webhook, path 'allow'",
			Webhooks: []registrationv1beta1.Webhook{{
				Name:              "cache5",
				ClientConfig:      ccfgURL("allow"),
				Rules:             newMatchEverythingRules(),
				NamespaceSelector: &metav1.LabelSelector{},
				FailurePolicy:     &policyIgnore,
			}},
			ExpectAllow:     true,
			ExpectCacheMiss: false,
		},
	}
}

// ccfgSVC returns a client config using the service reference mechanism.
func ccfgSVC(urlPath string) registrationv1beta1.WebhookClientConfig {
	return registrationv1beta1.WebhookClientConfig{
		Service: &registrationv1beta1.ServiceReference{
			Name:      "webhook-test",
			Namespace: "default",
			Path:      &urlPath,
		},
		CABundle: testcerts.CACert,
	}
}

func newMatchEverythingRules() []registrationv1beta1.RuleWithOperations {
	return []registrationv1beta1.RuleWithOperations{{
		Operations: []registrationv1beta1.OperationType{registrationv1beta1.OperationAll},
		Rule: registrationv1beta1.Rule{
			APIGroups:   []string{"*"},
			APIVersions: []string{"*"},
			Resources:   []string{"*/*"},
		},
	}}
}

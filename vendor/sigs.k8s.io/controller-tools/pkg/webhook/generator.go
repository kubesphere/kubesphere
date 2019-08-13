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

package webhook

import (
	"errors"
	"net"
	"net/url"
	"path"
	"sort"
	"strconv"

	"k8s.io/api/admissionregistration/v1beta1"
	admissionregistration "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apitypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

type generatorOptions struct {
	// webhooks maps a path to a webhoook.
	webhooks map[string]webhook

	// port is the port number that the server will serve.
	// It will be defaulted to 443 if unspecified.
	port int32

	// certDir is the directory that contains the server key and certificate.
	certDir string

	// mutatingWebhookConfigName is the name that used for creating the MutatingWebhookConfiguration object.
	mutatingWebhookConfigName string
	// validatingWebhookConfigName is the name that used for creating the ValidatingWebhookConfiguration object.
	validatingWebhookConfigName string

	// secret is the location for storing the certificate for the admission server.
	// The server should have permission to create a secret in the namespace.
	secret *apitypes.NamespacedName

	// service is a k8s service fronting the webhook server pod(s).
	// One and only one of service and host can be set.
	// This maps to field .Webhooks.ClientConfig.Service
	// https://github.com/kubernetes/api/blob/183f3326a9353bd6d41430fc80f96259331d029c/admissionregistration/v1beta1/types.go#L260
	service *service
	// host is the host name of .Webhooks.ClientConfig.URL
	// https://github.com/kubernetes/api/blob/183f3326a9353bd6d41430fc80f96259331d029c/admissionregistration/v1beta1/types.go#L250
	// One and only one of service and host can be set.
	// If neither service nor host is unspecified, host will be defaulted to "localhost".
	host *string
}

// service contains information for creating a Service
type service struct {
	// name of the Service
	name string
	// namespace of the Service
	namespace string
	// selectors is the selector of the Service.
	// This must select the pods that runs this webhook server.
	selectors map[string]string
}

// setDefaults does defaulting for the generatorOptions.
func (o *generatorOptions) setDefaults() {
	if o.webhooks == nil {
		o.webhooks = map[string]webhook{}
	}
	if o.port <= 0 {
		o.port = 443
	}
	if len(o.certDir) == 0 {
		o.certDir = path.Join("/tmp", "k8s-webhook-server", "serving-certs")
	}

	if len(o.mutatingWebhookConfigName) == 0 {
		o.mutatingWebhookConfigName = "mutating-webhook-configuration"
	}
	if len(o.validatingWebhookConfigName) == 0 {
		o.validatingWebhookConfigName = "validating-webhook-configuration"
	}
	if o.host == nil && o.service == nil {
		varString := "localhost"
		o.host = &varString
	}
}

// Generate creates the AdmissionWebhookConfiguration objects and Service if any.
// It also provisions the certificate for the admission server.
func (o *generatorOptions) Generate() ([]runtime.Object, error) {
	// do defaulting if necessary
	o.setDefaults()

	webhookConfigurations, err := o.whConfigs()
	if err != nil {
		return nil, err
	}
	svc := o.getService()
	objects := append(webhookConfigurations, svc)

	return objects, nil
}

// whConfigs creates a mutatingWebhookConfiguration and(or) a validatingWebhookConfiguration.
func (o *generatorOptions) whConfigs() ([]runtime.Object, error) {
	for _, webhook := range o.webhooks {
		if err := webhook.Validate(); err != nil {
			return nil, err
		}
	}

	objs := []runtime.Object{}
	mutatingWH, err := o.mutatingWHConfig()
	if err != nil {
		return nil, err
	}
	if mutatingWH != nil {
		objs = append(objs, mutatingWH)
	}
	validatingWH, err := o.validatingWHConfigs()
	if err != nil {
		return nil, err
	}
	if validatingWH != nil {
		objs = append(objs, validatingWH)
	}
	return objs, nil
}

// mutatingWHConfig creates mutatingWebhookConfiguration.
func (o *generatorOptions) mutatingWHConfig() (runtime.Object, error) {
	mutatingWebhooks := []v1beta1.Webhook{}
	for path, webhook := range o.webhooks {
		if webhook.GetType() != mutatingWebhook {
			continue
		}

		aw := webhook.(*admissionWebhook)
		wh, err := o.admissionWebhook(path, aw)
		if err != nil {
			return nil, err
		}
		mutatingWebhooks = append(mutatingWebhooks, *wh)
	}

	sort.Slice(mutatingWebhooks, func(i, j int) bool {
		return mutatingWebhooks[i].Name < mutatingWebhooks[j].Name
	})

	if len(mutatingWebhooks) > 0 {
		return &admissionregistration.MutatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: metav1.GroupVersion{Group: admissionregistration.GroupName, Version: "v1beta1"}.String(),
				Kind:       "MutatingWebhookConfiguration",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: o.mutatingWebhookConfigName,
				Annotations: map[string]string{
					// TODO(DirectXMan12): Change the annotation to the format that cert-manager decides to use.
					"alpha.admissionwebhook.cert-manager.io": "true",
				},
			},
			Webhooks: mutatingWebhooks,
		}, nil
	}
	return nil, nil
}

func (o *generatorOptions) validatingWHConfigs() (runtime.Object, error) {
	validatingWebhooks := []v1beta1.Webhook{}
	for path, webhook := range o.webhooks {
		var aw *admissionWebhook
		if webhook.GetType() != validatingWebhook {
			continue
		}

		aw = webhook.(*admissionWebhook)
		wh, err := o.admissionWebhook(path, aw)
		if err != nil {
			return nil, err
		}
		validatingWebhooks = append(validatingWebhooks, *wh)
	}

	sort.Slice(validatingWebhooks, func(i, j int) bool {
		return validatingWebhooks[i].Name < validatingWebhooks[j].Name
	})

	if len(validatingWebhooks) > 0 {
		return &admissionregistration.ValidatingWebhookConfiguration{
			TypeMeta: metav1.TypeMeta{
				APIVersion: metav1.GroupVersion{Group: admissionregistration.GroupName, Version: "v1beta1"}.String(),
				Kind:       "ValidatingWebhookConfiguration",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: o.validatingWebhookConfigName,
				Annotations: map[string]string{
					// TODO(DirectXMan12): Change the annotation to the format that cert-manager decides to use.
					"alpha.admissionwebhook.cert-manager.io": "true",
				},
			},
			Webhooks: validatingWebhooks,
		}, nil
	}
	return nil, nil
}

func (o *generatorOptions) admissionWebhook(path string, wh *admissionWebhook) (*admissionregistration.Webhook, error) {
	if wh.namespaceSelector == nil && o.service != nil && len(o.service.namespace) > 0 {
		wh.namespaceSelector = &metav1.LabelSelector{
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "control-plane",
					Operator: metav1.LabelSelectorOpDoesNotExist,
				},
			},
		}
	}

	webhook := &admissionregistration.Webhook{
		Name:              wh.name,
		Rules:             wh.rules,
		FailurePolicy:     wh.failurePolicy,
		NamespaceSelector: wh.namespaceSelector,
	}
	cc, err := o.getClientConfigWithPath(path)
	if err != nil {
		return nil, err
	}
	webhook.ClientConfig = *cc
	return webhook, nil
}

// getClientConfigWithPath constructs a WebhookClientConfig based on the server generatorOptions.
// It will use path to the set the path in WebhookClientConfig.
func (o *generatorOptions) getClientConfigWithPath(path string) (*admissionregistration.WebhookClientConfig, error) {
	cc, err := o.getClientConfig()
	if err != nil {
		return nil, err
	}
	return cc, setPath(cc, path)
}

func (o *generatorOptions) getClientConfig() (*admissionregistration.WebhookClientConfig, error) {
	if o.host != nil && o.service != nil {
		return nil, errors.New("URL and service can't be set at the same time")
	}
	cc := &admissionregistration.WebhookClientConfig{
		// Put an non-empty and not harmful CABundle here.
		// Not doing this will cause the field
		CABundle: []byte(`\n`),
	}
	if o.host != nil {
		u := url.URL{
			Scheme: "https",
			Host:   net.JoinHostPort(*o.host, strconv.Itoa(int(o.port))),
		}
		urlString := u.String()
		cc.URL = &urlString
	}
	if o.service != nil {
		cc.Service = &admissionregistration.ServiceReference{
			Name:      o.service.name,
			Namespace: o.service.namespace,
			// Path will be set later
		}
	}
	return cc, nil
}

// setPath sets the path in the WebhookClientConfig.
func setPath(cc *admissionregistration.WebhookClientConfig, path string) error {
	if cc.URL != nil {
		u, err := url.Parse(*cc.URL)
		if err != nil {
			return err
		}
		u.Path = path
		urlString := u.String()
		cc.URL = &urlString
	}
	if cc.Service != nil {
		cc.Service.Path = &path
	}
	return nil
}

// getService creates a corev1.Service object fronting the admission server.
func (o *generatorOptions) getService() runtime.Object {
	if o.service == nil {
		return nil
	}
	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      o.service.name,
			Namespace: o.service.namespace,
			Annotations: map[string]string{
				// Secret here only need name, since it will be in the same namespace as the service.
				// TODO(DirectXMan12): Change the annotation to the format that cert-manager decides to use.
				"alpha.service.cert-manager.io/serving-cert-secret-name": o.secret.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: o.service.selectors,
			Ports: []corev1.ServicePort{
				{
					// When using service, kube-apiserver will send admission request to port 443.
					Port:       443,
					TargetPort: intstr.IntOrString{Type: intstr.Int, IntVal: o.port},
				},
			},
		},
	}
	return svc
}

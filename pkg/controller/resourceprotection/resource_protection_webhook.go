/*
 * Copyright 2024 the KubeSphere Authors.
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
 */

package resourceprotection

import (
	"context"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"kubesphere.io/kubesphere/pkg/constants"
	kscontroller "kubesphere.io/kubesphere/pkg/controller"
)

const webhookName = "resource-protection-webhook"

func (w *Webhook) Name() string {
	return webhookName
}

type Webhook struct {
	client.Client
}

func (w *Webhook) SetupWithManager(mgr *kscontroller.Manager) error {
	w.Client = mgr.GetClient()
	mgr.GetWebhookServer().Register("/resource-protector", &webhook.Admission{Handler: w})
	return nil
}

func (w *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.Operation == admissionv1.Delete {
		gvr := req.RequestResource
		gvk, err := w.RESTMapper().KindFor(schema.GroupVersionResource{
			Group:    gvr.Group,
			Version:  gvr.Version,
			Resource: gvr.Resource,
		})
		if err != nil {
			return webhook.Errored(http.StatusInternalServerError, err)
		}
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(gvk)
		if err = w.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, obj); err != nil {
			return webhook.Errored(http.StatusInternalServerError, err)
		}

		if obj.GetLabels()[constants.ProtectedResourceLabel] == "true" {
			return webhook.Denied("this resource may not be deleted")
		}
	}
	return admission.Allowed("")
}

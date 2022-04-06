package webhooks

import (
	"context"

	accessor "github.com/kubesphere/storageclass-accessor/webhook"
	v1 "k8s.io/api/admission/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type AccessorHandler struct {
	C       client.Client
	decoder *admission.Decoder
}

func (h *AccessorHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

func (h *AccessorHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	review := v1.AdmissionReview{
		Request: &req.AdmissionRequest,
	}
	resp := accessor.AdmitPVC(review)
	return admission.Response{
		AdmissionResponse: *resp,
	}
}

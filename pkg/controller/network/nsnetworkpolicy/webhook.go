package nsnetworkpolicy

import (
	"context"
	"fmt"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:path=/validate-service-nsnp-kubesphere-io-v1alpha1-network,name=validate-v1-service,mutating=false,failurePolicy=fail,groups="",resources=services,verbs=create;update,versions=v1

// serviceValidator validates service
type ServiceValidator struct {
	decoder *admission.Decoder
}

// Service must hash label, becasue nsnp will use it
func (v *ServiceValidator) Handle(ctx context.Context, req admission.Request) admission.Response {
	service := &corev1.Service{}

	err := v.decoder.Decode(req, service)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	if service.Spec.Selector == nil {
		return admission.Denied(fmt.Sprintf("missing label"))
	}

	return admission.Allowed("")
}

func (a *ServiceValidator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}

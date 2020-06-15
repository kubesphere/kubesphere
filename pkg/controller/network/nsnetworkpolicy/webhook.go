// Copyright 2020 KubeSphere Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

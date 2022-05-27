// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
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

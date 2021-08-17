/*
Copyright 2020 KubeSphere Authors

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

package webhooks

import (
	"context"
	"net/http"
	"sync"

	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Validator defines functions for validating an operation
type Validator interface {
	ValidateCreate(obj runtime.Object) error
	ValidateUpdate(old runtime.Object, new runtime.Object) error
	ValidateDelete(obj runtime.Object) error
}

type ValidatorWrap struct {
	Obj    runtime.Object
	Helper Validator
}

type validators struct {
	vs   map[string]*ValidatorWrap
	lock sync.RWMutex
}

var (
	vs validators
)

func init() {
	vs = validators{
		vs:   make(map[string]*ValidatorWrap),
		lock: sync.RWMutex{},
	}
}

func RegisterValidator(name string, v *ValidatorWrap) {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	vs.vs[name] = v
}

func UnRegisterValidator(name string) {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	delete(vs.vs, name)
}

func GetValidator(name string) *ValidatorWrap {
	vs.lock.Lock()
	defer vs.lock.Unlock()

	return vs.vs[name]
}

type ValidatingHandler struct {
	C       client.Client
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &ValidatingHandler{}

// InjectDecoder injects the decoder into a ValidatingHandler.
func (h *ValidatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

// Handle handles admission requests.
func (h *ValidatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	validator := GetValidator(req.Kind.String())
	if validator == nil {
		return admission.Denied("crd has webhook configured, but the controller does not register the corresponding processing logic and refuses the operation by default.")
	}

	// Get the object in the request
	obj := validator.Obj.DeepCopyObject()
	if req.Operation == v1.Create {
		err := h.decoder.Decode(req, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = validator.Helper.ValidateCreate(obj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1.Update {
		oldObj := obj.DeepCopyObject()

		err := h.decoder.DecodeRaw(req.Object, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}
		err = h.decoder.DecodeRaw(req.OldObject, oldObj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = validator.Helper.ValidateUpdate(oldObj, obj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	if req.Operation == v1.Delete {
		// In reference to PR: https://github.com/kubernetes/kubernetes/pull/76346
		// OldObject contains the object being deleted
		err := h.decoder.DecodeRaw(req.OldObject, obj)
		if err != nil {
			return admission.Errored(http.StatusBadRequest, err)
		}

		err = validator.Helper.ValidateDelete(obj)
		if err != nil {
			return admission.Denied(err.Error())
		}
	}

	return admission.Allowed("")
}

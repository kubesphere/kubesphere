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

package admission

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-logr/logr"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"

	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
)

var (
	errUnableToEncodeResponse = errors.New("unable to encode response")
)

// Request defines the input for an admission handler.
// It contains information to identify the object in
// question (group, version, kind, resource, subresource,
// name, namespace), as well as the operation in question
// (e.g. Get, Create, etc), and the object itself.
type Request struct {
	admissionv1beta1.AdmissionRequest
}

// Response is the output of an admission handler.
// It contains a response indicating if a given
// operation is allowed, as well as a set of patches
// to mutate the object in the case of a mutating admission handler.
type Response struct {
	// Patches are the JSON patches for mutating webhooks.
	// Using this instead of setting Response.Patch to minimize
	// overhead of serialization and deserialization.
	// Patches set here will override any patches in the response,
	// so leave this empty if you want to set the patch response directly.
	Patches []jsonpatch.JsonPatchOperation
	// AdmissionResponse is the raw admission response.
	// The Patch field in it will be overwritten by the listed patches.
	admissionv1beta1.AdmissionResponse
}

// Complete populates any fields that are yet to be set in
// the underlying AdmissionResponse, It mutates the response.
func (r *Response) Complete(req Request) error {
	r.UID = req.UID

	// ensure that we have a valid status code
	if r.Result == nil {
		r.Result = &metav1.Status{}
	}
	if r.Result.Code == 0 {
		r.Result.Code = http.StatusOK
	}
	// TODO(directxman12): do we need to populate this further, and/or
	// is code actually necessary (the same webhook doesn't use it)

	if len(r.Patches) == 0 {
		return nil
	}

	var err error
	r.Patch, err = json.Marshal(r.Patches)
	if err != nil {
		return err
	}
	patchType := admissionv1beta1.PatchTypeJSONPatch
	r.PatchType = &patchType

	return nil
}

// Handler can handle an AdmissionRequest.
type Handler interface {
	Handle(context.Context, Request) Response
}

// HandlerFunc implements Handler interface using a single function.
type HandlerFunc func(context.Context, Request) Response

var _ Handler = HandlerFunc(nil)

// Handle process the AdmissionRequest by invoking the underlying function.
func (f HandlerFunc) Handle(ctx context.Context, req Request) Response {
	return f(ctx, req)
}

// Webhook represents each individual webhook.
type Webhook struct {
	// Handler actually processes an admission request returning whether it was allowed or denied,
	// and potentially patches to apply to the handler.
	Handler Handler

	// decoder is constructed on receiving a scheme and passed down to then handler
	decoder *Decoder

	log logr.Logger
}

// InjectLogger gets a handle to a logging instance, hopefully with more info about this particular webhook.
func (w *Webhook) InjectLogger(l logr.Logger) error {
	w.log = l
	return nil
}

// Handle processes AdmissionRequest.
// If the webhook is mutating type, it delegates the AdmissionRequest to each handler and merge the patches.
// If the webhook is validating type, it delegates the AdmissionRequest to each handler and
// deny the request if anyone denies.
func (w *Webhook) Handle(ctx context.Context, req Request) Response {
	resp := w.Handler.Handle(ctx, req)
	if err := resp.Complete(req); err != nil {
		w.log.Error(err, "unable to encode response")
		return Errored(http.StatusInternalServerError, errUnableToEncodeResponse)
	}

	return resp
}

// InjectScheme injects a scheme into the webhook, in order to construct a Decoder.
func (w *Webhook) InjectScheme(s *runtime.Scheme) error {
	// TODO(directxman12): we should have a better way to pass this down

	var err error
	w.decoder, err = NewDecoder(s)
	if err != nil {
		return err
	}

	// inject the decoder here too, just in case the order of calling this is not
	// scheme first, then inject func
	if w.Handler != nil {
		if _, err := InjectDecoderInto(w.GetDecoder(), w.Handler); err != nil {
			return err
		}
	}

	return nil
}

// GetDecoder returns a decoder to decode the objects embedded in admission requests.
// It may be nil if we haven't received a scheme to use to determine object types yet.
func (w *Webhook) GetDecoder() *Decoder {
	return w.decoder
}

// InjectFunc injects the field setter into the webhook.
func (w *Webhook) InjectFunc(f inject.Func) error {
	// inject directly into the handlers.  It would be more correct
	// to do this in a sync.Once in Handle (since we don't have some
	// other start/finalize-type method), but it's more efficient to
	// do it here, presumably.

	// also inject a decoder, and wrap this so that we get a setFields
	// that injects a decoder (hopefully things don't ignore the duplicate
	// InjectorInto call).

	var setFields inject.Func
	setFields = func(target interface{}) error {
		if err := f(target); err != nil {
			return err
		}

		if _, err := inject.InjectorInto(setFields, target); err != nil {
			return err
		}

		if _, err := InjectDecoderInto(w.GetDecoder(), target); err != nil {
			return err
		}

		return nil
	}

	return setFields(w.Handler)
}

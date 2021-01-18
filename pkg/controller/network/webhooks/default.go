package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Defaulter defines functions for setting defaults on resources
type Defaulter interface {
	Default(obj runtime.Object) error
}

type DefaulterWrap struct {
	Obj    runtime.Object
	Helper Defaulter
}

type MutatingHandler struct {
	C       client.Client
	decoder *admission.Decoder
}

var _ admission.DecoderInjector = &MutatingHandler{}

// InjectDecoder injects the decoder into a MutatingHandler.
func (h *MutatingHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}

type defaulters struct {
	ds   map[string]*DefaulterWrap
	lock sync.RWMutex
}

var (
	ds defaulters
)

func init() {
	ds = defaulters{
		ds:   make(map[string]*DefaulterWrap),
		lock: sync.RWMutex{},
	}
}

func RegisterDefaulter(name string, d *DefaulterWrap) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	ds.ds[name] = d
}

func UnRegisterDefaulter(name string) {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	delete(ds.ds, name)
}

func GetDefaulter(name string) *DefaulterWrap {
	ds.lock.Lock()
	defer ds.lock.Unlock()

	return ds.ds[name]
}

// Handle handles admission requests.
func (h *MutatingHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	defaulter := GetDefaulter(req.Kind.String())
	if defaulter == nil {
		return admission.Denied("crd has webhook configured, but the controller does not register the corresponding processing logic and refuses the operation by default.")
	}

	// Get the object in the request
	obj := defaulter.Obj.DeepCopyObject()
	err := h.decoder.Decode(req, obj)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	// Default the object
	defaulter.Helper.Default(obj)
	marshalled, err := json.Marshal(obj)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Create the patch
	return admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
}

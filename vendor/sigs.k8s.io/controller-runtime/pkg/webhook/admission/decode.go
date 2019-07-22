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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
)

// Decoder knows how to decode the contents of an admission
// request into a concrete object.
type Decoder struct {
	codecs serializer.CodecFactory
}

// NewDecoder creates a Decoder given the runtime.Scheme
func NewDecoder(scheme *runtime.Scheme) (*Decoder, error) {
	return &Decoder{codecs: serializer.NewCodecFactory(scheme)}, nil
}

// Decode decodes the inlined object in the AdmissionRequest into the passed-in runtime.Object.
// If you want decode the OldObject in the AdmissionRequest, use DecodeRaw.
func (d *Decoder) Decode(req Request, into runtime.Object) error {
	return d.DecodeRaw(req.Object, into)
}

// DecodeRaw decodes a RawExtension object into the passed-in runtime.Object.
func (d *Decoder) DecodeRaw(rawObj runtime.RawExtension, into runtime.Object) error {
	// NB(directxman12): there's a bug/weird interaction between decoders and
	// the API server where the API server doesn't send a GVK on the embedded
	// objects, which means the unstructured decoder refuses to decode.  It
	// also means we can't pass the unstructured directly in, since it'll try
	// and call unstructured's special Unmarshal implementation, which calls
	// back into that same decoder :-/
	// See kubernetes/kubernetes#74373.
	if unstructuredInto, isUnstructured := into.(*unstructured.Unstructured); isUnstructured {
		// unmarshal into unstructured's underlying object to avoid calling the decoder
		if err := json.Unmarshal(rawObj.Raw, &unstructuredInto.Object); err != nil {
			return err
		}

		return nil
	}

	deserializer := d.codecs.UniversalDeserializer()
	return runtime.DecodeInto(deserializer, rawObj.Raw, into)
}

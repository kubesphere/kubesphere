/*
Copyright 2017 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/apis/authentication"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AdmissionReview describes an admission review request/response.
type AdmissionReview struct {
	metav1.TypeMeta

	// Request describes the attributes for the admission request.
	// +optional
	Request *AdmissionRequest

	// Response describes the attributes for the admission response.
	// +optional
	Response *AdmissionResponse
}

// AdmissionRequest describes the admission.Attributes for the admission request.
type AdmissionRequest struct {
	// UID is an identifier for the individual request/response. It allows us to distinguish instances of requests which are
	// otherwise identical (parallel requests, requests when earlier requests did not modify etc)
	// The UID is meant to track the round trip (request/response) between the KAS and the WebHook, not the user request.
	// It is suitable for correlating log entries between the webhook and apiserver, for either auditing or debugging.
	UID types.UID
	// Kind is the type of object being manipulated.  For example: Pod
	Kind metav1.GroupVersionKind
	// Resource is the name of the resource being requested.  This is not the kind.  For example: pods
	Resource metav1.GroupVersionResource
	// SubResource is the name of the subresource being requested.  This is a different resource, scoped to the parent
	// resource, but it may have a different kind. For instance, /pods has the resource "pods" and the kind "Pod", while
	// /pods/foo/status has the resource "pods", the sub resource "status", and the kind "Pod" (because status operates on
	// pods). The binding resource for a pod though may be /pods/foo/binding, which has resource "pods", subresource
	// "binding", and kind "Binding".
	// +optional
	SubResource string
	// Name is the name of the object as presented in the request.  On a CREATE operation, the client may omit name and
	// rely on the server to generate the name.  If that is the case, this method will return the empty string.
	// +optional
	Name string
	// Namespace is the namespace associated with the request (if any).
	// +optional
	Namespace string
	// Operation is the operation being performed
	Operation Operation
	// UserInfo is information about the requesting user
	UserInfo authentication.UserInfo
	// Object is the object from the incoming request prior to default values being applied
	// +optional
	Object runtime.Object
	// OldObject is the existing object. Only populated for UPDATE requests.
	// +optional
	OldObject runtime.Object
}

// AdmissionResponse describes an admission response.
type AdmissionResponse struct {
	// UID is an identifier for the individual request/response.
	// This should be copied over from the corresponding AdmissionRequest.
	UID types.UID
	// Allowed indicates whether or not the admission request was permitted.
	Allowed bool
	// Result contains extra details into why an admission request was denied.
	// This field IS NOT consulted in any way if "Allowed" is "true".
	// +optional
	Result *metav1.Status
	// Patch contains the actual patch. Currently we only support a response in the form of JSONPatch, RFC 6902.
	// +optional
	Patch []byte
	// PatchType indicates the form the Patch will take. Currently we only support "JSONPatch".
	// +optional
	PatchType *PatchType
}

// PatchType is the type of patch being used to represent the mutated object
type PatchType string

// PatchType constants.
const (
	PatchTypeJSONPatch PatchType = "JSONPatch"
)

// Operation is the type of resource operation being checked for admission control
type Operation string

// Operation constants
const (
	Create  Operation = "CREATE"
	Update  Operation = "UPDATE"
	Delete  Operation = "DELETE"
	Connect Operation = "CONNECT"
)

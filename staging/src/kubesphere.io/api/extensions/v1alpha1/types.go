/*

 Copyright 2022 The KubeSphere Authors.

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
package v1alpha1

// ServiceReference holds a reference to Service.legacy.k8s.io
type ServiceReference struct {
	// namespace is the namespace of the service.
	// Required
	Namespace string `json:"namespace" protobuf:"bytes,1,opt,name=namespace"`
	// name is the name of the service.
	// Required
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`

	// path is an optional URL path at which the webhook will be contacted.
	// +optional
	Path *string `json:"path,omitempty" protobuf:"bytes,3,opt,name=path"`

	// port is an optional service port at which the webhook will be contacted.
	// `port` should be a valid port number (1-65535, inclusive).
	// Defaults to 443 for backward compatibility.
	// +optional
	Port *int32 `json:"port,omitempty" protobuf:"varint,4,opt,name=port"`
}

type Endpoint struct {
	// +optional
	URL *string `json:"url,omitempty"`
	// +optional
	Service *ServiceReference `json:"service,omitempty"`
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

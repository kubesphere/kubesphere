/*
Copyright 2018 The Kubernetes Authors
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

package component

import (
	"github.com/kubernetes-sigs/application/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Handle is an interface for operating on logical Components of a CR
type Handle interface {
	DependentResources(rsrc interface{}) *resource.ObjectBag
	ExpectedResources(rsrc interface{}, labels map[string]string, dependent, aggregated *resource.ObjectBag) (*resource.ObjectBag, error)
	Observables(scheme *runtime.Scheme, rsrc interface{}, labels map[string]string, expected *resource.ObjectBag) []resource.Observable
	Mutate(rsrc interface{}, rsrclabels map[string]string, status interface{}, expected, dependent, observed *resource.ObjectBag) (*resource.ObjectBag, error)
	Differs(expected metav1.Object, observed metav1.Object) bool
	UpdateComponentStatus(rsrc, status interface{}, reconciled *resource.ObjectBag, err error)
	Finalize(rsrc, status interface{}, observed *resource.ObjectBag) error
}

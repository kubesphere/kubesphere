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

package v1beta1

import (
	"github.com/kubernetes-sigs/application/pkg/component"
	cr "github.com/kubernetes-sigs/application/pkg/customresource"
	"github.com/kubernetes-sigs/application/pkg/finalizer"
	"github.com/kubernetes-sigs/application/pkg/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
)

// Mutate - mutate expected
func (a *Application) Mutate(rsrc interface{}, labels map[string]string, status interface{}, expected, dependent, observed *resource.ObjectBag) (*resource.ObjectBag, error) {
	exp := resource.ObjectBag{}
	for _, o := range observed.Items() {
		o.Lifecycle = resource.LifecycleManaged
		exp.Add(o)
	}
	return &exp, nil
}

// Finalize - execute finalizers
func (a *Application) Finalize(rsrc, sts interface{}, observed *resource.ObjectBag) error {
	finalizer.Remove(a, finalizer.Cleanup)
	return nil
}

//DependentResources - returns dependent rsrc
func (a *Application) DependentResources(rsrc interface{}) *resource.ObjectBag {
	return &resource.ObjectBag{}
}

// ExpectedResources returns the list of resource/name for those resources created by
// the operator for this spec and those resources referenced by this operator.
// Mark resources as owned, referred
func (a *Application) ExpectedResources(rsrc interface{}, rsrclabels map[string]string, dependent, aggregated *resource.ObjectBag) (*resource.ObjectBag, error) {
	return &resource.ObjectBag{}, nil
}

// GKVersions returns the gvks for the given gk
func GKVersions(s *runtime.Scheme, mgk metav1.GroupKind) []schema.GroupVersionKind {
	gvks := []schema.GroupVersionKind{}
	gk := schema.GroupKind{Group: mgk.Group, Kind: mgk.Kind}
	for gvk := range s.AllKnownTypes() {
		if gk != gvk.GroupKind() {
			continue
		}
		gvks = append(gvks, gvk)
	}
	return gvks
}

// Observables - return selectors
func (a *Application) Observables(scheme *runtime.Scheme, rsrc interface{}, rsrclabels map[string]string, expected *resource.ObjectBag) []resource.Observable {
	var observables []resource.Observable
	if a.Spec.Selector == nil || a.Spec.Selector.MatchLabels == nil {
		return observables
	}
	for _, gk := range a.Spec.ComponentGroupKinds {
		listGK := gk
		if !strings.HasSuffix(listGK.Kind, "List") {
			listGK.Kind = listGK.Kind + "List"
		}
		for _, gvk := range GKVersions(scheme, listGK) {
			ol, err := scheme.New(gvk)
			if err == nil {
				observable := resource.Observable{
					ObjList: ol.(metav1.ListInterface),
					Labels:  a.Spec.Selector.MatchLabels,
				}
				observables = append(observables, observable)
			}
		}

	}
	return observables
}

// Differs returns true if the resource needs to be updated
func (a *Application) Differs(expected metav1.Object, observed metav1.Object) bool {
	return false
}

// UpdateComponentStatus use reconciled objects to update component status
func (a *Application) UpdateComponentStatus(rsrci, statusi interface{}, reconciled *resource.ObjectBag, err error) {
	if a != nil {
		stts := statusi.(*ApplicationStatus)
		stts.UpdateStatus(reconciled.Objs(), err)
	}
}

// ApplyDefaults default app crd
func (a *Application) ApplyDefaults() {
	return
}

// UpdateRsrcStatus records status or error in status
func (a *Application) UpdateRsrcStatus(status interface{}, err error) bool {
	appstatus := status.(*ApplicationStatus)
	if status != nil {
		a.Status = *appstatus
	}

	if err != nil {
		a.Status.SetError("ErrorSeen", err.Error())
	} else {
		a.Status.ClearError()
	}
	return true
}

// Validate the Application
func (a *Application) Validate() error {
	return nil
}

// Components returns components for this resource
func (a *Application) Components() []component.Component {
	c := []component.Component{}
	c = append(c, component.Component{
		Handle:   a,
		Name:     "app",
		CR:       a,
		OwnerRef: a.OwnerRef(),
	})
	return c
}

// OwnerRef returns owner ref object with the component's resource as owner
func (a *Application) OwnerRef() *metav1.OwnerReference {
	if !a.Spec.AddOwnerRef {
		return nil
	}

	isController := false
	gvk := schema.GroupVersionKind{
		Group:   SchemeGroupVersion.Group,
		Version: SchemeGroupVersion.Version,
		Kind:    "Application",
	}
	ref := metav1.NewControllerRef(a, gvk)
	ref.Controller = &isController
	return ref
}

// NewRsrc - return a new resource object
func (a *Application) NewRsrc() cr.Handle {
	return &Application{}
}

// NewStatus - return a  resource status object
func (a *Application) NewStatus() interface{} {
	s := a.Status.DeepCopy()
	s.ComponentList = ComponentList{}
	return s
}

// StatusDiffers returns True if there is a change in status
func (a *Application) StatusDiffers(new ApplicationStatus) bool {
	return true
}

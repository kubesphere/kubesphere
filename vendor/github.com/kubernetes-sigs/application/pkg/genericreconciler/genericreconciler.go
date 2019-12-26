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

package genericreconciler

import (
	"context"
	"fmt"
	"github.com/kubernetes-sigs/application/pkg/component"
	cr "github.com/kubernetes-sigs/application/pkg/customresource"
	"github.com/kubernetes-sigs/application/pkg/resource"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	urt "k8s.io/apimachinery/pkg/util/runtime"
	"log"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func handleErrorArr(info string, name string, e error, errs []error) []error {
	HandleError(info, name, e)
	return append(errs, e)
}

// HandleError common error handling routine
func HandleError(info string, name string, e error) error {
	urt.HandleError(fmt.Errorf("Failed: [%s] %s. %s", name, info, e.Error()))
	return e
}

func (gr *Reconciler) observe(observables ...resource.Observable) (*resource.ObjectBag, error) {
	var returnval = new(resource.ObjectBag)
	var err error
	for _, obs := range observables {
		var resources []resource.Object
		if obs.Labels != nil {
			opts := client.ListOptions{
				LabelSelector: labels.SelectorFromSet(obs.Labels),
			}
			opts.Raw = &metav1.ListOptions{TypeMeta: obs.Type}
			err = gr.List(context.TODO(), obs.ObjList.(runtime.Object), &opts)
			if err == nil {
				items, err := meta.ExtractList(obs.ObjList.(runtime.Object))
				if err == nil {
					for _, item := range items {
						resources = append(resources, resource.Object{Obj: item.(metav1.Object)})
					}
				}
			}
		} else {
			var obj = obs.Obj.(metav1.Object)
			name := obj.GetName()
			namespace := obj.GetNamespace()
			otype := reflect.TypeOf(obj).String()
			err = gr.Get(context.TODO(),
				types.NamespacedName{Name: name, Namespace: namespace},
				obs.Obj.(runtime.Object))
			if err == nil {
				resources = append(resources, resource.Object{Obj: obs.Obj})
			} else {
				log.Printf("   >>>ERR get: %s", otype+"/"+namespace+"/"+name)
			}
		}
		if err != nil {
			return nil, err
		}
		for _, resource := range resources {
			returnval.Add(resource)
		}
	}
	return returnval, nil
}

func specDiffers(o1, o2 metav1.Object) bool {
	// Not all k8s objects have Spec
	// example ConfigMap
	// TODO strategic merge patch diff in generic controller loop
	e := reflect.Indirect(reflect.ValueOf(o1)).FieldByName("Spec")
	o := reflect.Indirect(reflect.ValueOf(o2)).FieldByName("Spec")
	if !e.IsValid() {
		// handling ConfigMap
		e = reflect.Indirect(reflect.ValueOf(o1)).FieldByName("Data")
		o = reflect.Indirect(reflect.ValueOf(o2)).FieldByName("Data")
	}

	if e.IsValid() && o.IsValid() {
		if reflect.DeepEqual(e.Interface(), o.Interface()) {
			return false
		}
	}
	return true
}

// If both ownerRefs have the same group/kind/name but different uid, that means at least one of them doesn't exist anymore.
// If we compare `uid` in this function, we'd set both as owners which is not what we want
// Because in the case that the original owner is already gone, we want its dependent to be garbage collected with it.
func isReferringSameObject(a, b metav1.OwnerReference) bool {
	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}
	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}
	return aGV == bGV && a.Kind == b.Kind && a.Name == b.Name
}

func injectOwnerRefs(o metav1.Object, ref *metav1.OwnerReference) bool {
	if ref == nil {
		return false
	}
	objRefs := o.GetOwnerReferences()
	for _, r := range objRefs {
		if isReferringSameObject(*ref, r) {
			return false
		}
	}
	objRefs = append(objRefs, *ref)
	o.SetOwnerReferences(objRefs)
	return true
}

// ReconcileCR is a generic function that reconciles expected and observed resources
func (gr *Reconciler) ReconcileCR(namespacedname types.NamespacedName, handle cr.Handle) error {
	var status interface{}
	expected := &resource.ObjectBag{}
	update := false
	rsrc := handle.NewRsrc()
	name := reflect.TypeOf(rsrc).String() + "/" + namespacedname.String()
	err := gr.Get(context.TODO(), namespacedname, rsrc.(runtime.Object))
	if err == nil {
		o := rsrc.(metav1.Object)
		err = rsrc.Validate()
		status = rsrc.NewStatus()
		if err == nil {
			rsrc.ApplyDefaults()
			components := rsrc.Components()
			for _, component := range components {
				if o.GetDeletionTimestamp() == nil {
					err = gr.ReconcileComponent(name, component, status, expected)
				} else {
					err = gr.FinalizeComponent(name, component, status, expected)
				}
			}
		}
	} else {
		if errors.IsNotFound(err) {
			urt.HandleError(fmt.Errorf("not found %s. %s", name, err.Error()))
			return nil
		}
	}
	update = rsrc.UpdateRsrcStatus(status, err)

	if update {
		err = gr.Update(context.TODO(), rsrc.(runtime.Object))
	}
	if err != nil {
		urt.HandleError(fmt.Errorf("error updating %s. %s", name, err.Error()))
	}

	return err
}

// ObserveAndMutate is a function that is called to observe and mutate expected resources
func (gr *Reconciler) ObserveAndMutate(crname string, c component.Component, status interface{}, mutate bool, aggregated *resource.ObjectBag) (*resource.ObjectBag, *resource.ObjectBag, error) {
	var err error
	var expected, observed, dependent *resource.ObjectBag
	emptybag := &resource.ObjectBag{}

	// Get dependenta objects
	dependent, err = gr.observe(resource.ObservablesFromObjects(gr.Scheme, c.DependentResources(c.CR), c.Labels())...)
	if err != nil {
		return emptybag, emptybag, fmt.Errorf("Failed getting dependent resources: %s", err.Error())
	}

	// Get Expected resources
	expected, err = c.ExpectedResources(c.CR, c.Labels(), dependent, aggregated)
	if err != nil {
		return emptybag, emptybag, fmt.Errorf("Failed gathering expected resources: %s", err.Error())
	}

	// Get observables
	observables := c.Observables(gr.Scheme, c.CR, c.Labels(), expected)

	// Observe observables
	observed, err = gr.observe(observables...)
	if err != nil {
		return emptybag, emptybag, fmt.Errorf("Failed observing resources: %s", err.Error())
	}

	// Mutate expected objects
	if mutate {
		expected, err = c.Mutate(c.CR, c.Labels(), status, expected, dependent, observed)
		if err != nil {
			return emptybag, emptybag, fmt.Errorf("Failed mutating resources: %s", err.Error())
		}

		// Get observables
		observables := c.Observables(gr.Scheme, c.CR, c.Labels(), expected)

		// Observe observables
		observed, err = gr.observe(observables...)
		if err != nil {
			return emptybag, emptybag, fmt.Errorf("Failed observing resources after mutation: %s", err.Error())
		}
	}

	return expected, observed, err
}

// FinalizeComponent is a function that finalizes component
func (gr *Reconciler) FinalizeComponent(crname string, c component.Component, status interface{}, aggregated *resource.ObjectBag) error {
	cname := crname + "(cmpnt:" + c.Name + ")"
	log.Printf("%s  finalizing component\n", cname)
	defer log.Printf("%s finalizing component completed", cname)

	expected, observed, err := gr.ObserveAndMutate(crname, c, status, false, aggregated)

	if err != nil {
		HandleError("", crname, err)
	}
	aggregated.Add(expected.Items()...)
	err = c.Finalize(c.CR, status, observed)
	return err
}

// ReconcileComponent is a generic function that reconciles expected and observed resources
func (gr *Reconciler) ReconcileComponent(crname string, c component.Component, status interface{}, aggregated *resource.ObjectBag) error {
	errs := []error{}
	var reconciled *resource.ObjectBag = new(resource.ObjectBag)

	cname := crname + "(cmpnt:" + c.Name + ")"
	log.Printf("%s  reconciling component\n", cname)
	defer log.Printf("%s  reconciling component completed\n", cname)

	expected, observed, err := gr.ObserveAndMutate(crname, c, status, true, aggregated)

	// Reconciliation logic is straight-forward:
	// This method gets the list of expected resources and observed resources
	// We compare the 2 lists and:
	//  create(rsrc) where rsrc is in expected but not in observed
	//  delete(rsrc) where rsrc is in observed but not in expected
	//  update(rsrc) where rsrc is in observed and expected
	//
	// We have a notion of Managed and Referred resources
	// Only Managed resources are CRUD'd
	// Missing Reffered resources are treated as errors and surfaced as such in the status field
	//

	if err != nil {
		errs = handleErrorArr("", crname, err, errs)
	} else {
		aggregated.Add(expected.Items()...)
	}

	for _, e := range expected.Items() {
		seen := false
		eNamespace := e.Obj.GetNamespace()
		eName := e.Obj.GetName()
		eKind := reflect.TypeOf(e.Obj).String()
		eRsrcInfo := eNamespace + "/" + eKind + "/" + eName
		for _, o := range observed.Items() {
			if (eName != o.Obj.GetName()) || (eNamespace != o.Obj.GetNamespace()) ||
				(eKind != reflect.TypeOf(o.Obj).String()) {
				continue
			}
			// rsrc is seen in both expected and observed, update it if needed
			e.Obj.SetResourceVersion(o.Obj.GetResourceVersion())
			e.Obj.SetOwnerReferences(o.Obj.GetOwnerReferences())
			if e.Lifecycle == resource.LifecycleManaged && (specDiffers(e.Obj, o.Obj) && c.Differs(e.Obj, o.Obj) || injectOwnerRefs(e.Obj, c.OwnerRef)) {
				if err := gr.Update(context.TODO(), e.Obj.(runtime.Object).DeepCopyObject()); err != nil {
					errs = handleErrorArr("update", eRsrcInfo, err, errs)
				}
			}
			reconciled.Add(o)
			seen = true
			break
		}
		// rsrc is in expected but not in observed - create
		if !seen {
			if e.Lifecycle == resource.LifecycleManaged {
				injectOwnerRefs(e.Obj, c.OwnerRef)
				if err := gr.Create(context.TODO(), e.Obj.(runtime.Object)); err != nil {
					errs = handleErrorArr("Create", cname, err, errs)
				} else {
					reconciled.Add(e)
				}
			} else {
				err := fmt.Errorf("missing resource not managed by %s: %s", cname, eRsrcInfo)
				errs = handleErrorArr("missing resource", cname, err, errs)
			}
		}
	}

	err = utilerrors.NewAggregate(errs)
	c.UpdateComponentStatus(c.CR, status, reconciled, err)
	return err
}

// Reconcile expected by kubebuilder
func (gr *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	err := gr.ReconcileCR(request.NamespacedName, gr.Handle)
	if err != nil {
		fmt.Printf("err: %s", err.Error())
	}
	return reconcile.Result{}, err
}

// AddToSchemes for adding Application to scheme
var AddToSchemes runtime.SchemeBuilder

// Init sets up Reconciler
func (gr *Reconciler) Init() {
	gr.Client = gr.Manager.GetClient()
	gr.Scheme = gr.Manager.GetScheme()
	AddToSchemes.AddToScheme(gr.Scheme)
}

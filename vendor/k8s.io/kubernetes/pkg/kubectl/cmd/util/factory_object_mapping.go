/*
Copyright 2016 The Kubernetes Authors.

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

// this file contains factories with no other dependencies

package util

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"sort"
	"sync"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/kubernetes/pkg/apis/apps"
	"k8s.io/kubernetes/pkg/apis/batch"
	api "k8s.io/kubernetes/pkg/apis/core"
	apiv1 "k8s.io/kubernetes/pkg/apis/core/v1"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/controller"
	"k8s.io/kubernetes/pkg/kubectl"
	"k8s.io/kubernetes/pkg/kubectl/categories"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	openapivalidation "k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi/validation"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/kubectl/validation"
	"k8s.io/kubernetes/pkg/printers"
	printersinternal "k8s.io/kubernetes/pkg/printers/internalversion"
)

type ring1Factory struct {
	clientAccessFactory ClientAccessFactory

	// openAPIGetter loads and caches openapi specs
	openAPIGetter openAPIGetter
}

type openAPIGetter struct {
	once   sync.Once
	getter openapi.Getter
}

func NewObjectMappingFactory(clientAccessFactory ClientAccessFactory) ObjectMappingFactory {
	f := &ring1Factory{
		clientAccessFactory: clientAccessFactory,
	}
	return f
}

// RESTMapper returns a mapper.
func (f *ring1Factory) RESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := f.clientAccessFactory.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	// allow conversion between typed and unstructured objects
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoveryClient)
	// TODO: should this also indicate it recognizes typed objects?
	expander := NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

func (f *ring1Factory) CategoryExpander() categories.CategoryExpander {
	legacyExpander := categories.LegacyCategoryExpander

	discoveryClient, err := f.clientAccessFactory.DiscoveryClient()
	if err == nil {
		// fallback is the legacy expander wrapped with discovery based filtering
		fallbackExpander, err := categories.NewDiscoveryFilteredExpander(legacyExpander, discoveryClient)
		CheckErr(err)

		// by default use the expander that discovers based on "categories" field from the API
		discoveryCategoryExpander, err := categories.NewDiscoveryCategoryExpander(fallbackExpander, discoveryClient)
		CheckErr(err)

		return discoveryCategoryExpander
	}

	return legacyExpander
}

func (f *ring1Factory) ClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	cfg, err := f.clientAccessFactory.ClientConfig()
	if err != nil {
		return nil, err
	}
	if err := setKubernetesDefaults(cfg); err != nil {
		return nil, err
	}
	gvk := mapping.GroupVersionKind
	switch gvk.Group {
	case api.GroupName:
		cfg.APIPath = "/api"
	default:
		cfg.APIPath = "/apis"
	}
	gv := gvk.GroupVersion()
	cfg.GroupVersion = &gv
	return restclient.RESTClientFor(cfg)
}

func (f *ring1Factory) UnstructuredClientForMapping(mapping *meta.RESTMapping) (resource.RESTClient, error) {
	cfg, err := f.clientAccessFactory.BareClientConfig()
	if err != nil {
		return nil, err
	}
	if err := restclient.SetKubernetesDefaults(cfg); err != nil {
		return nil, err
	}
	cfg.APIPath = "/apis"
	if mapping.GroupVersionKind.Group == api.GroupName {
		cfg.APIPath = "/api"
	}
	gv := mapping.GroupVersionKind.GroupVersion()
	cfg.ContentConfig = dynamic.ContentConfig()
	cfg.GroupVersion = &gv
	return restclient.RESTClientFor(cfg)
}

func (f *ring1Factory) Describer(mapping *meta.RESTMapping) (printers.Describer, error) {
	clientConfig, err := f.clientAccessFactory.ClientConfig()
	if err != nil {
		return nil, err
	}
	// try to get a describer
	if describer, ok := printersinternal.DescriberFor(mapping.GroupVersionKind.GroupKind(), clientConfig); ok {
		return describer, nil
	}
	// if this is a kind we don't have a describer for yet, go generic if possible
	if genericDescriber, genericErr := genericDescriber(f.clientAccessFactory, mapping); genericErr == nil {
		return genericDescriber, nil
	}
	// otherwise return an unregistered error
	return nil, fmt.Errorf("no description has been implemented for %s", mapping.GroupVersionKind.String())
}

// helper function to make a generic describer, or return an error
func genericDescriber(clientAccessFactory ClientAccessFactory, mapping *meta.RESTMapping) (printers.Describer, error) {
	clientConfig, err := clientAccessFactory.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientConfigCopy := *clientConfig
	clientConfigCopy.APIPath = dynamic.LegacyAPIPathResolverFunc(mapping.GroupVersionKind)
	gv := mapping.GroupVersionKind.GroupVersion()
	clientConfigCopy.GroupVersion = &gv

	// used to fetch the resource
	dynamicClient, err := dynamic.NewClient(&clientConfigCopy, gv)
	if err != nil {
		return nil, err
	}

	// used to get events for the resource
	clientSet, err := clientAccessFactory.ClientSet()
	if err != nil {
		return nil, err
	}
	eventsClient := clientSet.Core()

	return printersinternal.GenericDescriberFor(mapping, dynamicClient, eventsClient), nil
}

func (f *ring1Factory) LogsForObject(object, options runtime.Object, timeout time.Duration) (*restclient.Request, error) {
	clientset, err := f.clientAccessFactory.ClientSet()
	if err != nil {
		return nil, err
	}
	opts, ok := options.(*api.PodLogOptions)
	if !ok {
		return nil, errors.New("provided options object is not a PodLogOptions")
	}

	switch t := object.(type) {
	case *api.Pod:
		return clientset.Core().Pods(t.Namespace).GetLogs(t.Name, opts), nil
	case *corev1.Pod:
		return clientset.Core().Pods(t.Namespace).GetLogs(t.Name, opts), nil
	}

	namespace, selector, err := selectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot get the logs from %T: %v", object, err)
	}
	sortBy := func(pods []*v1.Pod) sort.Interface { return controller.ByLogging(pods) }
	pod, numPods, err := GetFirstPod(clientset.Core(), namespace, selector.String(), timeout, sortBy)
	if err != nil {
		return nil, err
	}
	if numPods > 1 {
		fmt.Fprintf(os.Stderr, "Found %v pods, using pod/%v\n", numPods, pod.Name)
	}
	return clientset.Core().Pods(pod.Namespace).GetLogs(pod.Name, opts), nil
}

func selectorsForObject(object runtime.Object) (namespace string, selector labels.Selector, err error) {
	switch t := object.(type) {
	case *extensions.ReplicaSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *extensionsv1beta1.ReplicaSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1.ReplicaSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta2.ReplicaSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}

	case *api.ReplicationController:
		namespace = t.Namespace
		selector = labels.SelectorFromSet(t.Spec.Selector)
	case *corev1.ReplicationController:
		namespace = t.Namespace
		selector = labels.SelectorFromSet(t.Spec.Selector)

	case *apps.StatefulSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1.StatefulSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta1.StatefulSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta2.StatefulSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}

	case *extensions.DaemonSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *extensionsv1beta1.DaemonSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1.DaemonSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta2.DaemonSet:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}

	case *extensions.Deployment:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *extensionsv1beta1.Deployment:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1.Deployment:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta1.Deployment:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *appsv1beta2.Deployment:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}

	case *batch.Job:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}
	case *batchv1.Job:
		namespace = t.Namespace
		selector, err = metav1.LabelSelectorAsSelector(t.Spec.Selector)
		if err != nil {
			return "", nil, fmt.Errorf("invalid label selector: %v", err)
		}

	case *api.Service:
		namespace = t.Namespace
		if t.Spec.Selector == nil || len(t.Spec.Selector) == 0 {
			return "", nil, fmt.Errorf("invalid service '%s': Service is defined without a selector", t.Name)
		}
		selector = labels.SelectorFromSet(t.Spec.Selector)
	case *corev1.Service:
		namespace = t.Namespace
		if t.Spec.Selector == nil || len(t.Spec.Selector) == 0 {
			return "", nil, fmt.Errorf("invalid service '%s': Service is defined without a selector", t.Name)
		}
		selector = labels.SelectorFromSet(t.Spec.Selector)

	default:
		return "", nil, fmt.Errorf("selector for %T not implemented", object)
	}

	return namespace, selector, nil
}

func (f *ring1Factory) HistoryViewer(mapping *meta.RESTMapping) (kubectl.HistoryViewer, error) {
	external, err := f.clientAccessFactory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	return kubectl.HistoryViewerFor(mapping.GroupVersionKind.GroupKind(), external)
}

func (f *ring1Factory) Rollbacker(mapping *meta.RESTMapping) (kubectl.Rollbacker, error) {
	external, err := f.clientAccessFactory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	return kubectl.RollbackerFor(mapping.GroupVersionKind.GroupKind(), external)
}

func (f *ring1Factory) StatusViewer(mapping *meta.RESTMapping) (kubectl.StatusViewer, error) {
	clientset, err := f.clientAccessFactory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	return kubectl.StatusViewerFor(mapping.GroupVersionKind.GroupKind(), clientset)
}

func (f *ring1Factory) ApproximatePodTemplateForObject(object runtime.Object) (*api.PodTemplateSpec, error) {
	switch t := object.(type) {
	case *api.Pod:
		return &api.PodTemplateSpec{
			ObjectMeta: t.ObjectMeta,
			Spec:       t.Spec,
		}, nil
	case *api.ReplicationController:
		return t.Spec.Template, nil
	case *extensions.ReplicaSet:
		return &t.Spec.Template, nil
	case *extensions.DaemonSet:
		return &t.Spec.Template, nil
	case *extensions.Deployment:
		return &t.Spec.Template, nil
	case *batch.Job:
		return &t.Spec.Template, nil
	}

	return nil, fmt.Errorf("unable to extract pod template from type %v", reflect.TypeOf(object))
}

func (f *ring1Factory) AttachablePodForObject(object runtime.Object, timeout time.Duration) (*api.Pod, error) {
	clientset, err := f.clientAccessFactory.ClientSet()
	if err != nil {
		return nil, err
	}

	switch t := object.(type) {
	case *api.Pod:
		return t, nil
	case *corev1.Pod:
		internalPod := &api.Pod{}
		err := apiv1.Convert_v1_Pod_To_core_Pod(t, internalPod, nil)
		return internalPod, err

	}

	namespace, selector, err := selectorsForObject(object)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %v", object, err)
	}
	sortBy := func(pods []*v1.Pod) sort.Interface { return sort.Reverse(controller.ActivePods(pods)) }
	pod, _, err := GetFirstPod(clientset.Core(), namespace, selector.String(), timeout, sortBy)
	return pod, err
}

func (f *ring1Factory) Validator(validate bool) (validation.Schema, error) {
	if !validate {
		return validation.NullSchema{}, nil
	}

	resources, err := f.OpenAPISchema()
	if err != nil {
		return nil, err
	}

	return validation.ConjunctiveSchema{
		openapivalidation.NewSchemaValidation(resources),
		validation.NoDoubleKeySchema{},
	}, nil
}

// OpenAPISchema returns metadata and structural information about Kubernetes object definitions.
func (f *ring1Factory) OpenAPISchema() (openapi.Resources, error) {
	discovery, err := f.clientAccessFactory.DiscoveryClient()
	if err != nil {
		return nil, err
	}

	// Lazily initialize the OpenAPIGetter once
	f.openAPIGetter.once.Do(func() {
		// Create the caching OpenAPIGetter
		f.openAPIGetter.getter = openapi.NewOpenAPIGetter(discovery)
	})

	// Delegate to the OpenAPIGetter
	return f.openAPIGetter.getter.Get()
}

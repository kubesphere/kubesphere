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

package testserver

import (
	"fmt"
	"time"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/scale"
)

const (
	noxuInstanceNum int64 = 9223372036854775807
)

//NewRandomNameCustomResourceDefinition generates a CRD with random name to avoid name conflict in e2e tests
func NewRandomNameCustomResourceDefinition(scope apiextensionsv1beta1.ResourceScope) *apiextensionsv1beta1.CustomResourceDefinition {
	// ensure the singular doesn't end in an s for now
	gName := names.SimpleNameGenerator.GenerateName("foo") + "a"
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: gName + "s.mygroup.example.com"},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "mygroup.example.com",
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   gName + "s",
				Singular: gName,
				Kind:     gName,
				ListKind: gName + "List",
			},
			Scope: scope,
		},
	}
}

func NewNoxuCustomResourceDefinition(scope apiextensionsv1beta1.ResourceScope) *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "noxus.mygroup.example.com"},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "mygroup.example.com",
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     "noxus",
				Singular:   "nonenglishnoxu",
				Kind:       "WishIHadChosenNoxu",
				ShortNames: []string{"foo", "bar", "abc", "def"},
				ListKind:   "NoxuItemList",
				Categories: []string{"all"},
			},
			Scope: scope,
		},
	}
}

func NewNoxuInstance(namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "mygroup.example.com/v1beta1",
			"kind":       "WishIHadChosenNoxu",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"content": map[string]interface{}{
				"key": "value",
			},
			"num": map[string]interface{}{
				"num1": noxuInstanceNum,
				"num2": 1000000,
			},
		},
	}
}

func NewNoxu2CustomResourceDefinition(scope apiextensionsv1beta1.ResourceScope) *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "noxus2.mygroup.example.com"},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "mygroup.example.com",
			Version: "v1alpha1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:     "noxus2",
				Singular:   "nonenglishnoxu2",
				Kind:       "WishIHadChosenNoxu2",
				ShortNames: []string{"foo", "bar", "abc", "def"},
				ListKind:   "Noxu2ItemList",
			},
			Scope: scope,
		},
	}
}

func NewCurletCustomResourceDefinition(scope apiextensionsv1beta1.ResourceScope) *apiextensionsv1beta1.CustomResourceDefinition {
	return &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "curlets.mygroup.example.com"},
		Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
			Group:   "mygroup.example.com",
			Version: "v1beta1",
			Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
				Plural:   "curlets",
				Singular: "curlet",
				Kind:     "Curlet",
				ListKind: "CurletList",
			},
			Scope: scope,
		},
	}
}

func NewCurletInstance(namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "mygroup.example.com/v1beta1",
			"kind":       "Curlet",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
			"content": map[string]interface{}{
				"key": "value",
			},
		},
	}
}

// CreateNewCustomResourceDefinitionWatchUnsafe creates the CRD and makes sure
// the apiextension apiserver has installed the CRD. But it's not safe to watch
// the created CR. Please call CreateNewCustomResourceDefinition if you need to
// watch the CR.
func CreateNewCustomResourceDefinitionWatchUnsafe(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface) error {
	_, err := apiExtensionsClient.ApiextensionsV1beta1().CustomResourceDefinitions().Create(crd)
	if err != nil {
		return err
	}

	// wait until the resource appears in discovery
	err = wait.PollImmediate(500*time.Millisecond, 30*time.Second, func() (bool, error) {
		resourceList, err := apiExtensionsClient.Discovery().ServerResourcesForGroupVersion(crd.Spec.Group + "/" + crd.Spec.Version)
		if err != nil {
			return false, nil
		}
		for _, resource := range resourceList.APIResources {
			if resource.Name == crd.Spec.Names.Plural {
				return true, nil
			}
		}
		return false, nil
	})

	return err
}

func CreateNewCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface, dynamicClientSet dynamic.DynamicInterface) error {
	err := CreateNewCustomResourceDefinitionWatchUnsafe(crd, apiExtensionsClient)
	if err != nil {
		return err
	}

	// This is only for a test.  We need the watch cache to have a resource version that works for the test.
	// When new REST storage is created, the storage cacher for the CR starts asynchronously.
	// REST API operations return like list use the RV of etcd, but the storage cacher's reflector's list
	// can get a different RV because etcd can be touched in between the initial list operation (if that's what you're doing first)
	// and the storage cache reflector starting.
	// Later, you can issue a watch with the REST apis list.RV and end up earlier than the storage cacher.
	// The general working model is that if you get a "resourceVersion too old" message, you re-list and rewatch.
	// For this test, we'll actually cycle, "list/watch/create/delete" until we get an RV from list that observes the create and not an error.
	// This way all the tests that are checking for watches don't have to worry about RV too old problems because crazy things *could* happen
	// before like the created RV could be too old to watch.
	var primingErr error
	wait.PollImmediate(500*time.Millisecond, 30*time.Second, func() (bool, error) {
		primingErr = checkForWatchCachePrimed(crd, dynamicClientSet)
		if primingErr == nil {
			return true, nil
		}
		return false, nil
	})
	if primingErr != nil {
		return primingErr
	}

	return nil
}

func checkForWatchCachePrimed(crd *apiextensionsv1beta1.CustomResourceDefinition, dynamicClientSet dynamic.DynamicInterface) error {
	ns := ""
	if crd.Spec.Scope != apiextensionsv1beta1.ClusterScoped {
		ns = "aval"
	}

	gvr := schema.GroupVersionResource{Group: crd.Spec.Group, Version: crd.Spec.Version, Resource: crd.Spec.Names.Plural}
	var resourceClient dynamic.DynamicResourceInterface
	if crd.Spec.Scope != apiextensionsv1beta1.ClusterScoped {
		resourceClient = dynamicClientSet.Resource(gvr).Namespace(ns)
	} else {
		resourceClient = dynamicClientSet.Resource(gvr)
	}

	initialList, err := resourceClient.List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	initialListListMeta, err := meta.ListAccessor(initialList)
	if err != nil {
		return err
	}

	instanceName := "setup-instance"
	instance := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": crd.Spec.Group + "/" + crd.Spec.Version,
			"kind":       crd.Spec.Names.Kind,
			"metadata": map[string]interface{}{
				"namespace": ns,
				"name":      instanceName,
			},
			"alpha":   "foo_123",
			"beta":    10,
			"gamma":   "bar",
			"delta":   "hello",
			"epsilon": "foobar",
		},
	}
	if _, err := resourceClient.Create(instance); err != nil {
		return err
	}
	// we created something, clean it up
	defer func() {
		resourceClient.Delete(instanceName, nil)
	}()

	noxuWatch, err := resourceClient.Watch(metav1.ListOptions{ResourceVersion: initialListListMeta.GetResourceVersion()})
	if err != nil {
		return err
	}
	defer noxuWatch.Stop()

	select {
	case watchEvent := <-noxuWatch.ResultChan():
		if watch.Added == watchEvent.Type {
			return nil
		}
		return fmt.Errorf("expected add, but got %#v", watchEvent)

	case <-time.After(5 * time.Second):
		return fmt.Errorf("gave up waiting for watch event")
	}
}

func DeleteCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface) error {
	if err := apiExtensionsClient.Apiextensions().CustomResourceDefinitions().Delete(crd.Name, nil); err != nil {
		return err
	}
	err := wait.PollImmediate(500*time.Millisecond, 30*time.Second, func() (bool, error) {
		groupResource, err := apiExtensionsClient.Discovery().ServerResourcesForGroupVersion(crd.Spec.Group + "/" + crd.Spec.Version)
		if err != nil {
			if errors.IsNotFound(err) {
				return true, nil

			}
			return false, err
		}
		for _, g := range groupResource.APIResources {
			if g.Name == crd.Spec.Names.Plural {
				return false, nil
			}
		}
		return true, nil
	})
	return err
}

func GetCustomResourceDefinition(crd *apiextensionsv1beta1.CustomResourceDefinition, apiExtensionsClient clientset.Interface) (*apiextensionsv1beta1.CustomResourceDefinition, error) {
	return apiExtensionsClient.Apiextensions().CustomResourceDefinitions().Get(crd.Name, metav1.GetOptions{})
}

func CreateNewScaleClient(crd *apiextensionsv1beta1.CustomResourceDefinition, config *rest.Config) (scale.ScalesGetter, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	groupResource, err := discoveryClient.ServerResourcesForGroupVersion(crd.Spec.Group + "/" + crd.Spec.Version)
	if err != nil {
		return nil, err
	}

	resources := []*discovery.APIGroupResources{
		{
			Group: metav1.APIGroup{
				Name: crd.Spec.Group,
				Versions: []metav1.GroupVersionForDiscovery{
					{Version: crd.Spec.Version},
				},
				PreferredVersion: metav1.GroupVersionForDiscovery{Version: crd.Spec.Version},
			},
			VersionedResources: map[string][]metav1.APIResource{
				crd.Spec.Version: groupResource.APIResources,
			},
		},
	}

	restMapper := discovery.NewRESTMapper(resources)
	resolver := scale.NewDiscoveryScaleKindResolver(discoveryClient)

	return scale.NewForConfig(config, restMapper, dynamic.LegacyAPIPathResolverFunc, resolver)
}

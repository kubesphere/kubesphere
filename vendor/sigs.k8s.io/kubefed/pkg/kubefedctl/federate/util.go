/*
Copyright 2019 The Kubernetes Authors.

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

package federate

import (
	"bufio"
	"io"
	"os"

	"github.com/pkg/errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"
	versionhelper "k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/yaml"

	"sigs.k8s.io/kubefed/pkg/apis/core/typeconfig"
	ctlutil "sigs.k8s.io/kubefed/pkg/controller/util"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/enable"
	"sigs.k8s.io/kubefed/pkg/kubefedctl/util"
)

func RemoveUnwantedFields(resource *unstructured.Unstructured) error {
	unstructured.RemoveNestedField(resource.Object, "apiVersion")
	unstructured.RemoveNestedField(resource.Object, "kind")
	unstructured.RemoveNestedField(resource.Object, "status")

	// All metadata fields save labels should be cleared. Other
	// metadata fields will be set by the system on creation or
	// subsequently by controllers.
	labels, _, err := unstructured.NestedMap(resource.Object, "metadata", "labels")
	if err != nil {
		return errors.Wrap(err, "Failed to retrieve metadata.labels")
	}
	unstructured.RemoveNestedField(resource.Object, "metadata")
	if len(labels) > 0 {
		err := unstructured.SetNestedMap(resource.Object, labels, "metadata", "labels")
		if err != nil {
			return errors.Wrap(err, "Failed to set metadata.labels")
		}
	}

	return nil
}

func SetBasicMetaFields(resource *unstructured.Unstructured, apiResource metav1.APIResource, name, namespace, generateName string) {
	resource.SetKind(apiResource.Kind)
	gv := schema.GroupVersion{Group: apiResource.Group, Version: apiResource.Version}
	resource.SetAPIVersion(gv.String())
	resource.SetName(name)
	if generateName != "" {
		resource.SetGenerateName(generateName)
	}
	if apiResource.Namespaced {
		resource.SetNamespace(namespace)
	}
}

func namespacedAPIResourceMap(config *rest.Config, skipAPIResourceNames []string) (map[string]metav1.APIResource, error) {
	apiResourceLists, err := enable.GetServerPreferredResources(config)
	if err != nil {
		return nil, err
	}

	apiResources := make(map[string]metav1.APIResource)
	for _, apiResourceList := range apiResourceLists {
		if len(apiResourceList.APIResources) == 0 {
			continue
		}

		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			return nil, errors.Wrap(err, "Error parsing GroupVersion")
		}

		group := gv.Group
		if apiResourceGroupMatchesSkipName(skipAPIResourceNames, group) {
			// A whole group is skipped by the user
			continue
		}

		if group == "extensions" {
			// The strategy involved to choose a Group higher in order for k8s core
			// APIs is to consider "extensions" as the outdated group [This seems to
			// be true for all k8s APIResources, so far]. For example if "deployments"
			// exists in "extensions" and "apps"; "deployments.apps" will be chosen.
			// This doesn't apply to events but events are listed in
			// controllerCreatedAPIResourceNames and so are skipped always.

			// Skipping this also assumes that "extensions" is not the only
			// group exposed for this resource on the API Server, which probably
			// is safe as "extensions" is deprecated.
			// TODO(irfanurrehman): Document this.
			continue
		}

		for _, apiResource := range apiResourceList.APIResources {
			if !apiResource.Namespaced || util.IsFederatedAPIResource(apiResource.Kind, group) ||
				apiResourceMatchesSkipName(apiResource, skipAPIResourceNames, group) {
				continue
			}

			// For all other resources (say CRDs) same kinds in different groups
			// are treated as individual types. If there happens to be an API Resource
			// which enables conversion and allows query of the same resource across
			// different groups, a specific group resource will have to be chosen by
			// the user using --skip-names to skip the not chosen one(s).
			// TODO(irfanurrehman): Document this.

			// The individual apiResources do not have the group and version set
			apiResource.Group = group
			apiResource.Version = gv.Version
			groupQualifiedName := typeconfig.GroupQualifiedName(apiResource)
			if previousAPIResource, ok := apiResources[groupQualifiedName]; ok {
				if versionhelper.CompareKubeAwareVersionStrings(gv.Version, previousAPIResource.Version) <= 0 {
					// The newer version is not latest keep the previous.
					continue
				}
			}

			apiResources[groupQualifiedName] = apiResource
		}
	}

	return apiResources, nil
}

func apiResourceGroupMatchesSkipName(skipAPIResourceNames []string, group string) bool {
	for _, name := range skipAPIResourceNames {
		if name == "" {
			continue
		}
		if name == group {
			return true
		}
	}
	return false
}

func apiResourceMatchesSkipName(apiResource metav1.APIResource, skipAPIResourceNames []string, group string) bool {
	names := append(controllerCreatedAPIResourceNames, skipAPIResourceNames...)
	for _, name := range names {
		if name == "" {
			continue
		}
		if enable.NameMatchesResource(name, apiResource, group) {
			return true
		}
	}
	return false
}

// resources stores a list of resources for an api type
type resources struct {
	// resource type information
	apiResource metav1.APIResource
	// resource list
	resources []*unstructured.Unstructured
}

func getResourcesInNamespace(config *rest.Config, namespace string, skipAPIResourceNames []string) ([]resources, error) {
	apiResources, err := namespacedAPIResourceMap(config, skipAPIResourceNames)
	if err != nil {
		return nil, err
	}

	resourcesInNamespace := []resources{}
	for _, apiResource := range apiResources {
		client, err := ctlutil.NewResourceClient(config, &apiResource)
		if err != nil {
			return nil, errors.Wrapf(err, "Error creating client for %s", apiResource.Kind)
		}

		resourceList, err := client.Resources(namespace).List(metav1.ListOptions{})
		if apierrors.IsNotFound(err) || resourceList == nil {
			continue
		}
		if err != nil {
			return nil, errors.Wrapf(err, "Error listing resources for %s", apiResource.Kind)
		}

		// It would be a waste of cycles to iterate through empty slices while federating resource
		if len(resourceList.Items) == 0 {
			continue
		}

		targetResources := resources{apiResource: apiResource}
		for _, item := range resourceList.Items {
			resource := item
			errors := validation.IsDNS1123Subdomain(resource.GetName())
			if len(errors) == 0 {
				targetResources.resources = append(targetResources.resources, &resource)
			} else {
				klog.Warningf("Skipping resource %s of type %s because it does not conform to the DNS-1123 subdomain spec.", resource.GetName(), apiResource.Name)
				klog.Warningf("The following error(s) were reported during DNS-1123 validation: ")
				for _, err := range errors {
					klog.Warningf(err)
				}
			}
		}
		resourcesInNamespace = append(resourcesInNamespace, targetResources)
	}

	return resourcesInNamespace, nil
}

// decodeUnstructuredFromFile reads a list of yamls into a slice of unstructured objects
func DecodeUnstructuredFromFile(filename string) ([]*unstructured.Unstructured, error) {
	var f *os.File
	if filename == "-" {
		f = os.Stdin
	} else {
		var err error
		f, err = os.Open(filename)

		if err != nil {
			return nil, err
		}
	}
	defer f.Close()

	var unstructuredList []*unstructured.Unstructured
	reader := utilyaml.NewYAMLReader(bufio.NewReader(f))
	for {
		unstructuedObj := &unstructured.Unstructured{}
		// Read one YAML document at a time, until io.EOF is returned
		buf, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(buf) == 0 {
			break
		}
		if err := yaml.Unmarshal(buf, unstructuedObj); err != nil {
			return nil, err
		}

		unstructuredList = append(unstructuredList, unstructuedObj)
	}

	return unstructuredList, nil
}

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

package util

import (
	"io"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func WriteUnstructuredToYaml(unstructuredObj *unstructured.Unstructured, w io.Writer) error {
	// If status is included in the yaml, attempting to create it in a
	// kube API will cause an error.
	obj := unstructuredObj.DeepCopy()
	unstructured.RemoveNestedField(obj.Object, "status")
	unstructured.RemoveNestedField(obj.Object, "metadata", "creationTimestamp")

	errMsg := "Error encoding unstructured object to yaml"
	objJSON, err := obj.MarshalJSON()
	if err != nil {
		return errors.Wrap(err, errMsg)
	}

	data, err := yaml.JSONToYAML(objJSON)
	if err != nil {
		return errors.Wrap(err, errMsg)
	}
	_, err = w.Write(data)
	if err != nil {
		return errors.Wrap(err, errMsg)
	}
	return nil
}

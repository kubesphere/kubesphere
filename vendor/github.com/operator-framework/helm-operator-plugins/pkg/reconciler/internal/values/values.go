/*
Copyright 2020 The Operator-SDK Authors.

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

package values

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/strvals"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/operator-framework/helm-operator-plugins/pkg/values"
)

type Values struct {
	m map[string]interface{}
}

func FromUnstructured(obj *unstructured.Unstructured) (*Values, error) {
	if obj == nil || obj.Object == nil {
		return nil, fmt.Errorf("nil object")
	}
	spec, ok := obj.Object["spec"]
	if !ok {
		return nil, fmt.Errorf("spec not found")
	}
	specMap, ok := spec.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("spec must be a map")
	}
	return New(specMap), nil
}

func New(m map[string]interface{}) *Values {
	return &Values{m: m}
}

func (v *Values) Map() map[string]interface{} {
	if v == nil {
		return nil
	}
	return v.m
}

func (v *Values) ApplyOverrides(in map[string]string) error {
	for inK, inV := range in {
		val := fmt.Sprintf("%s=%s", inK, os.ExpandEnv(inV))
		if err := strvals.ParseInto(val, v.m); err != nil {
			return err
		}
	}
	return nil
}

var DefaultMapper = values.MapperFunc(func(v chartutil.Values) chartutil.Values { return v })

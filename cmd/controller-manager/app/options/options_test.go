// Copyright 2022 The KubeSphere Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ref: https://github.com/kubernetes/kubernetes/blob/master/staging/src/k8s.io/controller-manager/app/helper_test.go
func TestIsControllerEnabled(t *testing.T) {
	testcases := []struct {
		name            string
		controllerName  string
		controllerFlags []string
		expected        bool
	}{
		{
			name:            "on by name",
			controllerName:  "bravo",
			controllerFlags: []string{"alpha", "bravo", "-charlie"},
			expected:        true,
		},
		{
			name:            "off by name",
			controllerName:  "charlie",
			controllerFlags: []string{"alpha", "bravo", "-charlie"},
			expected:        false,
		},
		{
			name:            "on by default",
			controllerName:  "alpha",
			controllerFlags: []string{"*"},
			expected:        true,
		},
		{
			name:            "on by star, not off by name",
			controllerName:  "alpha",
			controllerFlags: []string{"*", "-charlie"},
			expected:        true,
		},
		{
			name:            "off by name with star",
			controllerName:  "charlie",
			controllerFlags: []string{"*", "-charlie"},
			expected:        false,
		},
		{
			name:            "off then on",
			controllerName:  "alpha",
			controllerFlags: []string{"-alpha", "alpha"},
			expected:        false,
		},
		{
			name:            "on then off",
			controllerName:  "alpha",
			controllerFlags: []string{"alpha", "-alpha"},
			expected:        true,
		},
	}

	for _, tc := range testcases {
		option := NewKubeSphereControllerManagerOptions()
		option.ControllerGates = tc.controllerFlags
		actual := option.IsControllerEnabled(tc.controllerName)
		assert.Equal(t, tc.expected, actual, "%v: expected %v, got %v", tc.name, tc.expected, actual)
	}
}

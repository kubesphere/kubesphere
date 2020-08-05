/*
Copyright 2020 KubeSphere Authors

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

package expressions

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestLabelReplace(t *testing.T) {
	tests := []struct {
		expr        string
		expected    string
		expectedErr bool
	}{
		{
			expr:        "up",
			expected:    `up{namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `up{namespace="random"}`,
			expected:    `up{namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `up{namespace="random"} + up{job="test"}`,
			expected:    `up{namespace="default"} + up{job="test",namespace="default"}`,
			expectedErr: false,
		},
		{
			expr:        `@@@@`,
			expectedErr: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			result, err := labelReplace(tt.expr, "default")
			if err != nil {
				if !tt.expectedErr {
					t.Fatal(err)
				}
				return
			}

			if diff := cmp.Diff(result, tt.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", tt.expected, diff)
			}
		})
	}
}

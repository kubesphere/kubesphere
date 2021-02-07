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

package sliceutil

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAppendArg(t *testing.T) {
	var tests = []struct {
		args     []string
		arg      string
		expected []string
	}{
		{
			args:     []string{"--arg1=val1", "--arg2=val2", "--arg3=val3"},
			arg:      "--arg4=val4",
			expected: []string{"--arg1=val1", "--arg2=val2", "--arg3=val3", "--arg4=val4"},
		},
		{
			args:     []string{"--arg1=val1", "--arg2=val2", "--arg3=val3"},
			arg:      "  --arg2=val2 ",
			expected: []string{"--arg1=val1", "--arg2=val2", "--arg3=val3"},
		},
		{
			args:     []string{"  --arg1 = val1", " --arg2=val2 ", " --arg3"},
			arg:      "  --arg2=val2 ",
			expected: []string{"  --arg1 = val1", " --arg2=val2 ", " --arg3"},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			result := AppendArg(test.args, test.arg)
			if diff := cmp.Diff(result, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

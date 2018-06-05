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

package get

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	api "k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/printers"
)

func TestHumanReadablePrinterSupportsExpectedOptions(t *testing.T) {
	testObject := &api.Pod{ObjectMeta: metav1.ObjectMeta{
		Name: "foo",
		Labels: map[string]string{
			"l1": "value",
		},
	}}

	testCases := []struct {
		name       string
		showKind   bool
		showLabels bool

		// TODO(juanvallejo): test sorting once it's moved to the HumanReadablePrinter
		sortBy       string
		columnLabels []string

		noHeaders     bool
		withNamespace bool

		outputFormat string

		expectedError  string
		expectedOutput string
		expectNoMatch  bool
	}{
		{
			name:           "empty output format matches a humanreadable printer",
			expectedOutput: "NAME\\ +READY\\ +STATUS\\ +RESTARTS\\ +AGE\nfoo\\ +0/0\\ +0\\ +<unknown>\n",
		},
		{
			name:           "\"wide\" output format prints",
			outputFormat:   "wide",
			expectedOutput: "NAME\\ +READY\\ +STATUS\\ +RESTARTS\\ +AGE\\ +IP\\ +NODE\nfoo\\ +0/0\\ +0\\ +<unknown>\\ +<none>\\ +<none>\n",
		},
		{
			name:           "no-headers prints output with no headers",
			noHeaders:      true,
			expectedOutput: "foo\\ +0/0\\ +0\\ +<unknown>\n",
		},
		{
			name:           "no-headers and a \"wide\" output format prints output with no headers and additional columns",
			outputFormat:   "wide",
			noHeaders:      true,
			expectedOutput: "foo\\ +0/0\\ +0\\ +<unknown>\\ +<none>\\ +<none>\n",
		},
		{
			name:           "show-kind displays the resource's kind, even when printing a single type of resource",
			showKind:       true,
			expectedOutput: "NAME\\ +READY\\ +STATUS\\ +RESTARTS\\ +AGE\npod/foo\\ +0/0\\ +0\\ +<unknown>\n",
		},
		{
			name:           "label-columns prints specified label values in new column",
			columnLabels:   []string{"l1"},
			expectedOutput: "NAME\\ +READY\\ +STATUS\\ +RESTARTS\\ +AGE\\ +L1\nfoo\\ +0/0\\ +0\\ +<unknown>\\ +value\n",
		},
		{
			name:           "withNamespace displays an additional NAMESPACE column",
			withNamespace:  true,
			expectedOutput: "NAMESPACE\\ +NAME\\ +READY\\ +STATUS\\ +RESTARTS\\ +AGE\n\\ +foo\\ +0/0\\ +0\\ +<unknown>\n",
		},
		{
			name:          "no printer is matched on an invalid outputFormat",
			outputFormat:  "invalid",
			expectNoMatch: true,
		},
		{
			name:          "printer should not match on any other format supported by another printer",
			outputFormat:  "go-template",
			expectNoMatch: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			printFlags := HumanPrintFlags{
				ShowKind:     &tc.showKind,
				ShowLabels:   &tc.showLabels,
				SortBy:       &tc.sortBy,
				ColumnLabels: &tc.columnLabels,

				NoHeaders:     tc.noHeaders,
				WithNamespace: tc.withNamespace,
			}

			if tc.showKind {
				printFlags.Kind = schema.GroupKind{Kind: "pod"}
			}

			p, err := printFlags.ToPrinter(tc.outputFormat)
			if tc.expectNoMatch {
				if !printers.IsNoCompatiblePrinterError(err) {
					t.Fatalf("expected no printer matches for output format %q", tc.outputFormat)
				}
				return
			}
			if printers.IsNoCompatiblePrinterError(err) {
				t.Fatalf("expected to match template printer for output format %q", tc.outputFormat)
			}

			if len(tc.expectedError) > 0 {
				if err == nil || !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("expecting error %q, got %v", tc.expectedError, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			out := bytes.NewBuffer([]byte{})
			err = p.PrintObj(testObject, out)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			match, err := regexp.Match(tc.expectedOutput, out.Bytes())
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !match {
				t.Errorf("unexpected output: expecting %q, got %q", tc.expectedOutput, out.String())
			}
		})
	}
}

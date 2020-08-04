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

package monitoring

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"
	"math"
	"testing"
)

func TestSort(t *testing.T) {
	tests := []struct {
		target     string
		order      string
		identifier string
		raw        string
		expected   string
	}{
		{
			target:     "node_cpu_utilisation",
			order:      "asc",
			identifier: "node",
			raw:        "source-node-metrics.json",
			expected:   "sorted-node-metrics-asc.json",
		},
		{
			target:     "node_memory_utilisation",
			order:      "desc",
			identifier: "node",
			raw:        "source-node-metrics.json",
			expected:   "sorted-node-metrics-desc.json",
		},
		{
			target:     "node_memory_utilisation",
			order:      "desc",
			identifier: "node",
			raw:        "faulty-node-metrics-1.json",
			expected:   "faulty-node-metrics-sorted-1.json",
		},
		{
			target:     "node_cpu_utilisation",
			order:      "asc",
			identifier: "node",
			raw:        "faulty-node-metrics-2.json",
			expected:   "faulty-node-metrics-sorted-2.json",
		},
		{
			target:     "node_cpu_utilisation",
			order:      "asc",
			identifier: "node",
			raw:        "faulty-node-metrics-3.json",
			expected:   "faulty-node-metrics-sorted-3.json",
		},
		{
			target:     "node_memory_utilisation",
			order:      "desc",
			identifier: "node",
			raw:        "blank-node-metrics.json",
			expected:   "blank-node-metrics-sorted.json",
		},
		{
			target:     "node_memory_utilisation",
			order:      "desc",
			identifier: "node",
			raw:        "null-node-metrics.json",
			expected:   "null-node-metrics-sorted.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			source, expected, err := jsonFromFile(tt.raw, tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			result := source.Sort(tt.target, tt.order, tt.identifier)
			opt := cmp.Comparer(func(x, y float64) bool {
				return (math.IsNaN(x) && math.IsNaN(y)) || x == y
			})
			if diff := cmp.Diff(*result, *expected, opt); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestPage(t *testing.T) {
	tests := []struct {
		page     int
		limit    int
		raw      string
		expected string
	}{
		{
			page:     0,
			limit:    5,
			raw:      "sorted-node-metrics-asc.json",
			expected: "sorted-node-metrics-asc.json",
		},
		{
			page:     1,
			limit:    5,
			raw:      "sorted-node-metrics-asc.json",
			expected: "paged-node-metrics-1.json",
		},
		{
			page:     2,
			limit:    5,
			raw:      "sorted-node-metrics-asc.json",
			expected: "paged-node-metrics-2.json",
		},
		{
			page:     3,
			limit:    5,
			raw:      "sorted-node-metrics-asc.json",
			expected: "paged-node-metrics-3.json",
		},
		{
			page:     1,
			limit:    2,
			raw:      "faulty-node-metrics-sorted-1.json",
			expected: "faulty-node-metrics-paged.json",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			source, expected, err := jsonFromFile(tt.raw, tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			result := source.Page(tt.page, tt.limit)
			if diff := cmp.Diff(*result, *expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func jsonFromFile(sourceFile, expectedFile string) (*Metrics, *Metrics, error) {
	sourceJson := &Metrics{}
	expectedJson := &Metrics{}

	json, err := ioutil.ReadFile(fmt.Sprintf("./testdata/%s", sourceFile))
	if err != nil {
		return nil, nil, err
	}
	err = jsoniter.Unmarshal(json, sourceJson)
	if err != nil {
		return nil, nil, err
	}

	json, err = ioutil.ReadFile(fmt.Sprintf("./testdata/%s", expectedFile))
	if err != nil {
		return nil, nil, err
	}
	err = jsoniter.Unmarshal(json, expectedJson)
	if err != nil {
		return nil, nil, err
	}
	return sourceJson, expectedJson, nil
}

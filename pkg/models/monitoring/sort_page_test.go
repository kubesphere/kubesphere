package monitoring

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"io/ioutil"
	"testing"
)

func TestSort(t *testing.T) {
	tests := []struct {
		name       string
		target     string
		order      string
		identifier string
		source     string
		expected   string
	}{
		{"sort in ascending order", "node_cpu_utilisation", "asc", "node", "source-node-metrics.json", "sorted-node-metrics-asc.json"},
		{"sort in descending order", "node_memory_utilisation", "desc", "node", "source-node-metrics.json", "sorted-node-metrics-desc.json"},
		{"sort faulty metrics", "node_memory_utilisation", "desc", "node", "faulty-node-metrics.json", "faulty-node-metrics-sorted.json"},
		{"sort metrics with an blank node", "node_memory_utilisation", "desc", "node", "blank-node-metrics.json", "blank-node-metrics-sorted.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, expected, err := jsonFromFile(tt.source, tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			result := source.Sort(tt.target, tt.order, tt.identifier)
			if diff := cmp.Diff(*result, *expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestPage(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		limit    int
		source   string
		expected string
	}{
		{"page 0 limit 5", 0, 5, "sorted-node-metrics-asc.json", "sorted-node-metrics-asc.json"},
		{"page 1 limit 5", 1, 5, "sorted-node-metrics-asc.json", "paged-node-metrics-1.json"},
		{"page 2 limit 5", 2, 5, "sorted-node-metrics-asc.json", "paged-node-metrics-2.json"},
		{"page 3 limit 5", 3, 5, "sorted-node-metrics-asc.json", "paged-node-metrics-3.json"},
		{"page faulty metrics", 1, 2, "faulty-node-metrics-sorted.json", "faulty-node-metrics-paged.json"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, expected, err := jsonFromFile(tt.source, tt.expected)
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

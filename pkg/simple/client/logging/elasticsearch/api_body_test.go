package elasticsearch

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"kubesphere.io/kubesphere/pkg/simple/client/logging"
	"reflect"
	"testing"
	"time"
)

func TestMainBool(t *testing.T) {
	var tests = []struct {
		filter   logging.SearchFilter
		expected string
	}{
		{
			filter: logging.SearchFilter{
				NamespaceFilter: map[string]time.Time{
					"default": time.Unix(1589981934, 0),
				},
			},
			expected: "api_body_1.json",
		},
		{
			filter: logging.SearchFilter{
				WorkloadFilter: []string{"mysql"},
				Starttime:      time.Unix(1589980934, 0),
				Endtime:        time.Unix(1589981934, 0),
			},
			expected: "api_body_2.json",
		},
		{
			filter: logging.SearchFilter{
				PodFilter: []string{"mysql"},
				PodSearch: []string{"mysql-a8w3s-10945j"},
				LogSearch: []string{"info"},
			},
			expected: "api_body_3.json",
		},
		{
			filter: logging.SearchFilter{
				ContainerFilter: []string{"mysql-1"},
				ContainerSearch: []string{"mysql-3"},
			},
			expected: "api_body_4.json",
		},
		{
			filter: logging.SearchFilter{
				Starttime: time.Unix(1590744676, 0),
			},
			expected: "api_body_7.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected Body
			err := JsonFromFile(test.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			result := newBodyBuilder().mainBool(test.filter).Body

			if diff := cmp.Diff(result, expected); diff != "" {
				fmt.Printf("%T differ (-got, +want): %s", expected, diff)
			}
		})
	}
}

func TestCardinalityAggregation(t *testing.T) {
	var tests = []struct {
		expected string
	}{
		{
			expected: "api_body_5.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected Body
			err := JsonFromFile(test.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			result := newBodyBuilder().cardinalityAggregation().Body

			if !reflect.DeepEqual(result, expected) {
				t.Fatalf("expected: %v, but got %v", expected, result)
			}
		})
	}
}

func TestDateHistogramAggregation(t *testing.T) {
	var tests = []struct {
		expected string
	}{
		{
			expected: "api_body_6.json",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			var expected Body
			err := JsonFromFile(test.expected, &expected)
			if err != nil {
				t.Fatal(err)
			}

			result := newBodyBuilder().dateHistogramAggregation("15m").Body

			if !reflect.DeepEqual(result, expected) {
				t.Fatalf("expected: %v, but got %v", expected, result)
			}
		})
	}
}

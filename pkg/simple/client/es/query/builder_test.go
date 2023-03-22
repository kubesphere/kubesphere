package query

import (
	"fmt"
	"reflect"
	"testing"
)

func TestNewTerms(t *testing.T) {
	var tests = []struct {
		fakeKey  string
		fakeVal  interface{}
		expected *Terms
	}{
		{
			fakeKey: "key1",
			fakeVal: "value",
			expected: &Terms{
				Terms: map[string]interface{}{
					"key1": "value",
				},
			},
		},
		{
			fakeKey: "key2",
			fakeVal: 0.33,
			expected: &Terms{
				Terms: map[string]interface{}{
					"key2": 0.33,
				},
			},
		},
		{
			fakeKey: "key3",
			fakeVal: []int32{1, 2, 3},
			expected: &Terms{
				Terms: map[string]interface{}{
					"key3": []int32{1, 2, 3},
				},
			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			terms := NewTerms(test.fakeKey, test.fakeVal)
			if !reflect.DeepEqual(terms, test.expected) {
				t.Fatalf("NewTerms() got=%v, want %v", terms, test.expected)
			}
		})
	}
}

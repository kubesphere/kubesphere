package sliceutil

import (
	"reflect"
	"sort"
	"testing"
)

func Test_StringDiff(t *testing.T) {
	inputs := [][][]string{
		[][]string{
			[]string{
				"a", "b", "c",
			},
			[]string{},
		},
		[][]string{
			[]string{},
			[]string{
				"a", "b", "c",
			},
		},
		[][]string{
			[]string{
				"a", "b", "c",
			},
			[]string{
				"a", "b", "c",
			},
		},
		[][]string{
			[]string{
				"1", "a", "b", "c", "d", "e",
			},
			[]string{
				"a", "b", "c",
			},
		},
		[][]string{
			[]string{
				"a", "b", "c",
			},
			[]string{
				"1", "a", "b", "c", "d", "e",
			},
		},
		[][]string{
			[]string{
				"a", "d", "cssss",
			},
			[]string{
				"cssss", "b", "c",
			},
		},
	}
	results := [][][]string{
		[][]string{
			[]string{},
			[]string{
				"a", "b", "c",
			},
		},
		[][]string{
			[]string{
				"a", "b", "c",
			},
			[]string{},
		},
		[][]string{
			[]string{},
			[]string{},
		},
		[][]string{
			[]string{},
			[]string{
				"1", "d", "e",
			},
		},
		[][]string{
			[]string{
				"1", "d", "e",
			},
			[]string{},
		},
		[][]string{
			[]string{
				"b", "c",
			},
			[]string{
				"a", "d",
			},
		},
	}

	for i, input := range inputs {
		more, less := StringDiff(input[0], input[1])
		sort.Strings(more)
		sort.Strings(less)
		sort.Strings(results[i][0])
		sort.Strings(results[i][1])

		if !reflect.DeepEqual(more, results[i][0]) {
			t.Fatalf("%v and %v more should be %v, but returns %v", input[0], input[1], results[i][0], more)
		}
		if !reflect.DeepEqual(less, results[i][1]) {
			t.Fatalf("%v and %v less should be %v, but returns %v", input[0], input[1], results[i][1], less)
		}
	}

}

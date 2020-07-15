package esutil

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"testing"
	"time"
)

func TestResolveIndexNames(t *testing.T) {
	var tests = []struct {
		prefix   string
		start    time.Time
		end      time.Time
		expected string
	}{
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			expected: "ks-logstash-log-2020.02.01",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			expected: "ks-logstash-log-2020.02.03,ks-logstash-log-2020.02.02,ks-logstash-log-2020.02.01",
		},
		{
			prefix:   "ks-logstash-log",
			end:      time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			expected: "ks-logstash-log*",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			expected: "ks-logstash-log*",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.February, 3, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2020, time.February, 1, 0, 0, 0, 0, time.UTC),
			expected: "",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			result := ResolveIndexNames(test.prefix, test.start, test.end)
			if diff := cmp.Diff(result, test.expected); diff != "" {
				t.Fatalf("%T differ (-got, +want): %s", test.expected, diff)
			}
		})
	}
}

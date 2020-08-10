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
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.August, 6, 22, 0, 0, 0, time.UTC),
			end:      time.Date(2020, time.August, 7, 04, 0, 0, 0, time.UTC),
			expected: "ks-logstash-log-2020.08.07,ks-logstash-log-2020.08.06",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.August, 6, 22, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			end:      time.Date(2020, time.August, 7, 04, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			expected: "ks-logstash-log-2020.08.06",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.August, 7, 02, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			end:      time.Date(2020, time.August, 7, 04, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			expected: "ks-logstash-log-2020.08.06",
		},
		{
			prefix:   "ks-logstash-log",
			start:    time.Date(2020, time.August, 7, 12, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			end:      time.Date(2020, time.August, 7, 14, 0, 0, 0, time.FixedZone("UTC+8", 8*3600)),
			expected: "ks-logstash-log-2020.08.07",
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

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

// TODO: refactor
package esutil

import (
	"fmt"
	"strings"
	"time"
)

// KubeSphere uses the layout `yyyy.MM.dd`.
const layoutISO = "2006.01.02"

// Always do calculation based on UTC.
func ResolveIndexNames(prefix string, start, end time.Time) string {
	if end.IsZero() {
		end = time.Now()
	}

	// In case of no start time or a broad query range over 30 days, search all indices.
	if start.IsZero() || end.Sub(start).Hours() > 24*30 {
		return fmt.Sprintf("%s*", prefix)
	}

	var indices []string
	days := int(end.Sub(start).Hours() / 24)
	if start.Add(time.Duration(days)*24*time.Hour).UTC().Day() != end.UTC().Day() {
		days++
	}
	for i := 0; i <= days; i++ {
		suffix := end.Add(time.Duration(-i) * 24 * time.Hour).UTC().Format(layoutISO)
		indices = append(indices, fmt.Sprintf("%s-%s", prefix, suffix))
	}

	// If start is after end, ResolveIndexNames returns "".
	// However, query parameter validation will prevent it from happening at the very beginning (Bad Request).
	return strings.Join(indices, ",")
}

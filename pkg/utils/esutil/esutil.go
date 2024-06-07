/*
 * Please refer to the LICENSE file in the root directory of the project.
 * https://github.com/kubesphere/kubesphere/blob/master/LICENSE
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

// TODO: refactor
package esutil

import (
	"fmt"
	"strings"
	"time"
)

// KubeSphere uses the layout `yyyy.MM.dd`.
const layoutISO = "2006.01.02"

func ResolveIndexNames(prefix string, start, end time.Time) string {
	if end.IsZero() {
		end = time.Now()
	}

	// In case of no start time or a broad query range over 30 days, search all indices.
	if start.IsZero() || end.Sub(start).Hours() > 24*30 {
		return fmt.Sprintf("%s*", prefix)
	}

	var indices []string
	for i := 0; i <= int(end.Sub(start).Hours()/24); i++ {
		suffix := end.Add(time.Duration(-i) * 24 * time.Hour).Format(layoutISO)
		indices = append(indices, fmt.Sprintf("%s-%s", prefix, suffix))
	}

	// If start is after end, ResolveIndexNames returns "".
	// However, query parameter validation will prevent it from happening at the very beginning (Bad Request).
	return strings.Join(indices, ",")
}

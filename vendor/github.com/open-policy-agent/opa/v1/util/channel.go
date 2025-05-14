package util

import (
	"github.com/open-policy-agent/opa/v1/metrics"
)

// This prevents getting blocked forever writing to a full buffer, in case another routine fills the last space.
// Retrying maxEventRetry times to drop the oldest event. Dropping the incoming event if there still isn't room.
const maxEventRetry = 1000

// PushFIFO pushes data into a buffered channel without blocking when full, making room by dropping the oldest data.
// An optional metric can be recorded when data is dropped.
func PushFIFO[T any](buffer chan T, data T, metrics metrics.Metrics, metricName string) {

	for range maxEventRetry {
		// non-blocking send to the buffer, to prevent blocking if buffer is full so room can be made.
		select {
		case buffer <- data:
			return
		default:
		}

		// non-blocking drop from the buffer to make room for incoming event
		select {
		case <-buffer:
			if metrics != nil && metricName != "" {
				metrics.Counter(metricName).Incr()
			}
		default:
		}
	}
}

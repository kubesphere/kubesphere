package util

import "time"

// TimerWithCancel exists because of memory leaks when using
// time.After in select statements. Instead, we now manually create timers,
// wait on them, and manually free them.
//
// See this for more details:
// https://www.arangodb.com/2020/09/a-story-of-a-memory-leak-in-go-how-to-properly-use-time-after/
//
// Note: This issue is fixed in Go 1.23, but this fix helps us until then.
//
// Warning: the cancel cannot be done concurrent to reading, everything should
// work in the same goroutine.
//
// Example:
//
//	for retries := 0; true; retries++ {
//
//		...main logic...
//
//		timer, cancel := utils.TimerWithCancel(utils.Backoff(retries))
//		select {
//		case <-ctx.Done():
//			cancel()
//			return ctx.Err()
//		case <-timer.C:
//			continue
//		}
//	}
func TimerWithCancel(delay time.Duration) (*time.Timer, func()) {
	timer := time.NewTimer(delay)

	return timer, func() {
		// Note: The Stop function returns:
		// - true: if the timer is active. (no draining required)
		// - false: if the timer was already stopped or fired/expired.
		// In this case the channel should be drained to prevent memory
		// leaks only if it is not empty.
		// This operation is safe only if the cancel function is
		// used in same goroutine. Concurrent reading or canceling may
		// cause deadlock.
		if !timer.Stop() && len(timer.C) > 0 {
			<-timer.C
		}
	}
}

// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package util

import (
	"math/rand"
	"time"
)

// DefaultBackoff returns a delay with an exponential backoff based on the
// number of retries.
func DefaultBackoff(base, maxNS float64, retries int) time.Duration {
	return Backoff(base, maxNS, .2, 1.6, retries)
}

// Backoff returns a delay with an exponential backoff based on the number of
// retries. Same algorithm used in gRPC.
func Backoff(base, maxNS, jitter, factor float64, retries int) time.Duration {
	if retries == 0 {
		return 0
	}

	backoff, maxNS := base, maxNS
	for backoff < maxNS && retries > 0 {
		backoff *= factor
		retries--
	}
	if backoff > maxNS {
		backoff = maxNS
	}

	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + jitter*(rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}

	return time.Duration(backoff)
}

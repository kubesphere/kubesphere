// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package util

import (
	"math/rand"
	"time"
)

func init() {
	// NOTE(sr): We don't need good random numbers here; it's used for jittering
	// the backup timing a bit. But anyways, let's make it random enough; without
	// a call to rand.Seed() we'd get the same stream of numbers for each program
	// run. (Or not, if some other packages happens to seed the global randomness
	// source.)
	// Note(philipc): rand.Seed() was deprecated in Go 1.20, so we've switched to
	// using the recommended rand.New(rand.NewSource(seed)) style.
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// DefaultBackoff returns a delay with an exponential backoff based on the
// number of retries.
func DefaultBackoff(base, max float64, retries int) time.Duration {
	return Backoff(base, max, .2, 1.6, retries)
}

// Backoff returns a delay with an exponential backoff based on the number of
// retries. Same algorithm used in gRPC.
func Backoff(base, max, jitter, factor float64, retries int) time.Duration {
	if retries == 0 {
		return 0
	}

	backoff, max := base, max
	for backoff < max && retries > 0 {
		backoff *= factor
		retries--
	}
	if backoff > max {
		backoff = max
	}

	// Randomize backoff delays so that if a cluster of requests start at
	// the same time, they won't operate in lockstep.
	backoff *= 1 + jitter*(rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}

	return time.Duration(backoff)
}

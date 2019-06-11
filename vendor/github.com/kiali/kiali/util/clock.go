package util

import "time"

type TimeProvider interface {
	Now() time.Time
}

var Clock TimeProvider

type RealClock struct{}

func (clock RealClock) Now() time.Time {
	return time.Now()
}

type ClockMock struct {
	Time time.Time
}

func (clock ClockMock) Now() time.Time {
	return clock.Time
}

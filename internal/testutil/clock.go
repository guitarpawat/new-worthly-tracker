package testutil

import "time"

type StaticClock struct {
	Time time.Time
}

func (c StaticClock) Now() time.Time {
	return c.Time
}

package util

import "time"

// ExponentialBackoff returns function that while sleep for a time
// doubling it each time it's called
func ExponentialBackoff(duration time.Duration) func() {
	d := duration
	return func() {
		time.Sleep(d)
		d = d * 2
	}
}

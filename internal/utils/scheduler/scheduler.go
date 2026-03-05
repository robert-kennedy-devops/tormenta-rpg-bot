package scheduler

import "time"

// Every runs fn in a goroutine on each ticker tick.
func Every(interval time.Duration, fn func()) {
	if interval <= 0 || fn == nil {
		return
	}
	go func() {
		t := time.NewTicker(interval)
		defer t.Stop()
		for range t.C {
			fn()
		}
	}()
}

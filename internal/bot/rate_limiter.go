package bot

import (
	"time"
)

// RateLimiter is a simple token-based limiter for Telegram API calls.
// It is intentionally lightweight and dependency-free for incremental adoption.
type RateLimiter struct {
	tokens chan struct{}
}

// NewRateLimiter creates a limiter with `ratePerSec` steady refill and `burst`.
func NewRateLimiter(ratePerSec int, burst int) *RateLimiter {
	if ratePerSec < 1 {
		ratePerSec = 1
	}
	if burst < 1 {
		burst = 1
	}

	rl := &RateLimiter{
		tokens: make(chan struct{}, burst),
	}

	for i := 0; i < burst; i++ {
		rl.tokens <- struct{}{}
	}

	interval := time.Second / time.Duration(ratePerSec)
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			select {
			case rl.tokens <- struct{}{}:
			default:
			}
		}
	}()

	return rl
}

// Wait blocks until one request token is available.
func (r *RateLimiter) Wait() {
	<-r.tokens
}

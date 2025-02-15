package ratelimiter

import "time"

type RateLimiter struct {
	MaxRequest int
	BlockTime  time.Duration
}

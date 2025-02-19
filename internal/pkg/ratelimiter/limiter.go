package ratelimiter

import "time"

type RateLimiter struct {
	MaxRequest int
	BlockTime  time.Duration
}

func NewRateLimiter(maxRequest int, blockTime time.Duration) *RateLimiter {
	return &RateLimiter{
		MaxRequest: maxRequest,
		BlockTime:  blockTime,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	return true
}

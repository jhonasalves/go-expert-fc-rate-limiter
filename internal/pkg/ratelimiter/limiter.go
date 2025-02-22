package ratelimiter

import (
	"context"
	"time"
)

type Options struct {
	MaxRequest int
	BlockTime  time.Duration
}

type RateLimiter struct {
	storage Storage
	opt     Options
}

func NewRateLimiter(storage Storage, opt Options) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		opt:     opt,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string) bool {
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return false
	}

	if blocked {
		return false
	}

	count, err := rl.storage.IncrRequest(ctx, key, rl.opt.BlockTime)
	if err != nil {
		return false
	}

	if count > rl.opt.MaxRequest {
		rl.storage.BlockRequest(ctx, key, rl.opt.BlockTime)
		return false
	}

	return true
}

package ratelimiter

import (
	"context"
	"time"
)

type RateLimiterResponse struct {
	Allowed      bool      `json:"allowed"`
	RetryAfter   time.Time `json:"retry_after,omitempty"`
	RequestsLeft int       `json:"requests_left"`
	Limit        int       `json:"limit"`
}

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

func (rl *RateLimiter) Allow(ctx context.Context, key string) (RateLimiterResponse, error) {
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	if blocked {
		return RateLimiterResponse{
			Allowed:      false,
			RetryAfter:   time.Now().Add(rl.opt.BlockTime),
			RequestsLeft: 0,
			Limit:        rl.opt.MaxRequest,
		}, nil
	}

	count, err := rl.storage.IncrRequest(ctx, key, rl.opt.BlockTime)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	if count > rl.opt.MaxRequest {
		rl.storage.BlockRequest(ctx, key, rl.opt.BlockTime)
		return RateLimiterResponse{
			Allowed:      false,
			RetryAfter:   time.Now().Add(rl.opt.BlockTime),
			RequestsLeft: 0,
			Limit:        rl.opt.MaxRequest,
		}, nil
	}

	return RateLimiterResponse{
		Allowed:      true,
		RetryAfter:   time.Time{},
		RequestsLeft: rl.opt.MaxRequest - count,
		Limit:        rl.opt.MaxRequest,
	}, nil
}

package ratelimiter

import (
	"context"
	"log/slog"
	"time"
)

type KeyType int

const (
	Token KeyType = iota
	API
)

type RateLimiterResponse struct {
	Allowed      bool      `json:"allowed"`
	RetryAfter   time.Time `json:"retry_after,omitempty"`
	RequestsLeft int       `json:"requests_left"`
	Limit        int       `json:"limit"`
}

type Options struct {
	MaxRequestIP    int
	MaxRequestToken int
	WindowDuration  time.Duration
	BlockDuration   time.Duration
}

type RateLimiter struct {
	storage Storage
	opts    Options
	logger  *slog.Logger
}

func NewRateLimiter(storage Storage, opts Options, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		storage: storage,
		opts:    opts,
		logger:  logger,
	}
}

func (rl *RateLimiter) Allow(ctx context.Context, key string, keyType KeyType) (RateLimiterResponse, error) {
	blocked, err := rl.storage.IsBlocked(ctx, key)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	if blocked {
		return RateLimiterResponse{
			Allowed:      false,
			RetryAfter:   time.Now().Add(rl.opts.BlockDuration),
			RequestsLeft: 0,
			Limit:        rl.opts.MaxRequestIP,
		}, nil
	}

	count, err := rl.storage.IncrRequest(ctx, key, rl.opts.BlockDuration)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	if count > rl.opts.MaxRequestIP {
		rl.storage.BlockRequest(ctx, key, rl.opts.BlockDuration)
		return RateLimiterResponse{
			Allowed:      false,
			RetryAfter:   time.Now().Add(rl.opts.BlockDuration),
			RequestsLeft: 0,
			Limit:        rl.opts.MaxRequestIP,
		}, nil
	}

	return RateLimiterResponse{
		Allowed:      true,
		RetryAfter:   time.Time{},
		RequestsLeft: rl.opts.MaxRequestIP - count,
		Limit:        rl.opts.MaxRequestIP,
	}, nil
}

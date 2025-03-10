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

type RateLimitKey struct {
	Key     string
	KeyType KeyType
}

type RateLimiterResponse struct {
	Allowed      bool      `json:"allowed"`
	ResetTime    time.Time `json:"reset_time,omitempty"`
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

func (rl *RateLimiter) Allow(ctx context.Context, rk RateLimitKey) (RateLimiterResponse, error) {
	blocked, retryAfter, err := rl.storage.IsBlocked(ctx, rk.Key)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	maxRequest := rl.getMaxRequest(rk)

	if blocked {
		return RateLimiterResponse{
			Allowed:      false,
			RetryAfter:   time.Now().Add(retryAfter),
			RequestsLeft: 0,
			Limit:        maxRequest,
		}, nil
	}

	count, resetTime, err := rl.storage.IncrRequest(ctx, rk.Key, rl.opts.WindowDuration)
	if err != nil {
		return RateLimiterResponse{}, err
	}

	if count > maxRequest {
		rl.storage.BlockRequest(ctx, rk.Key, rl.opts.BlockDuration)

		return RateLimiterResponse{
			Allowed:      false,
			ResetTime:    time.Now().Add(resetTime),
			RequestsLeft: 0,
			Limit:        maxRequest,
		}, nil
	}

	return RateLimiterResponse{
		Allowed:      true,
		ResetTime:    time.Now().Add(resetTime),
		RetryAfter:   time.Time{},
		RequestsLeft: maxRequest - count,
		Limit:        maxRequest,
	}, nil
}

func (rl *RateLimiter) getMaxRequest(rk RateLimitKey) int {
	if rk.KeyType == Token {
		return rl.opts.MaxRequestToken
	}

	return rl.opts.MaxRequestIP
}

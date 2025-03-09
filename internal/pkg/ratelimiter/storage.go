package ratelimiter

import (
	"context"
	"time"
)

type Storage interface {
	IncrRequest(ctx context.Context, key string, window time.Duration) (int, time.Duration, error)
	IsBlocked(ctx context.Context, key string) (bool, time.Duration, error)
	BlockRequest(ctx context.Context, key string, duration time.Duration) error
}

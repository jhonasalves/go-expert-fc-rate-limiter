package ratelimiter

import (
	"context"
	"time"
)

type Storage interface {
	IncrRequest(ctx context.Context, key string, window time.Duration) (int, error)
	GetRequest(ctx context.Context, key string) (int, error)
	ResetRequest(ctx context.Context, key string) error
	IsBlocked(ctx context.Context, key string) (bool, error)
	BlockRequest(ctx context.Context, key string, window time.Duration) error
}

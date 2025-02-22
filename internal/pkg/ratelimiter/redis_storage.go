package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const RateLimitPrefix = "rate_limiter:"

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(client *redis.Client) *RedisStorage {
	return &RedisStorage{
		client: client,
	}
}

func (r *RedisStorage) IncrRequest(ctx context.Context, key string, window time.Duration) (int, error) {
	count := r.client.Incr(ctx, key)

	if count.Err() != nil {
		return 0, count.Err()
	}

	if count.Val() == 1 {
		r.client.Expire(ctx, key, window)
	}

	return int(count.Val()), nil
}

func (r *RedisStorage) GetRequest(ctx context.Context, key string) (int, error) {
	value := r.client.Get(ctx, key)

	if value.Err() != nil {
		return 0, value.Err()
	}

	count, err := value.Int()
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (r *RedisStorage) ResetRequest(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	return r.client.Exists(ctx, key).Val() == 1, nil
}

func (r *RedisStorage) Block(ctx context.Context, key string, window time.Duration) error {
	return r.client.Set(ctx, key, 1, window).Err()
}

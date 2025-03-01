package ratelimiter

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const RateLimitPrefix = "rate_limiter:"

type RedisStorage struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisStorage(client *redis.Client, logger *slog.Logger) *RedisStorage {
	return &RedisStorage{
		client: client,
		logger: logger,
	}
}

func (r *RedisStorage) IncrRequest(ctx context.Context, key string, window time.Duration) (int, error) {
	count := r.client.Incr(ctx, key)

	r.logger.Info("Incrementing request count",
		slog.String("key", key),
		slog.Int("count", int(count.Val())),
	)

	if count.Err() != nil {
		r.logger.Error("Error incrementing request count",
			slog.String("key", key),
			slog.String("error", count.Err().Error()),
		)
		return 0, count.Err()
	}

	if count.Val() == 1 {
		r.logger.Info("Setting expiration for key",
			slog.String("key", key),
			slog.String("window", window.String()),
		)

		r.client.Expire(ctx, key, window)
	}

	return int(count.Val()), nil
}

func (r *RedisStorage) GetRequest(ctx context.Context, key string) (int, error) {
	value := r.client.Get(ctx, key)

	r.logger.Info("Getting request count",
		slog.String("key", key),
		slog.String("value", value.String()),
	)

	if value.Err() != nil {
		r.logger.Error("Error getting request count",
			slog.String("key", key),
			slog.String("error", value.Err().Error()),
		)

		return 0, value.Err()
	}

	count, err := value.Int()
	if err != nil {
		r.logger.Error("Error converting value to int",
			slog.String("key", key),
			slog.String("error", err.Error()),
		)
		return 0, err
	}

	r.logger.Info("Returning request count",
		slog.String("key", key),
		slog.Int("count", count),
	)

	return int(count), nil
}

func (r *RedisStorage) ResetRequest(ctx context.Context, key string) error {
	r.logger.Info("Deleting key",
		slog.String("key", key),
	)

	return r.client.Del(ctx, key).Err()
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
	r.logger.Info("Checking if key is blocked",
		slog.String("key", key),
	)

	return r.client.Exists(ctx, key).Val() == 1, nil
}

func (r *RedisStorage) BlockRequest(ctx context.Context, key string, window time.Duration) error {
	r.logger.Info("Setting key expiration",
		slog.String("key", key),
		slog.String("window", window.String()),
	)

	return r.client.Set(ctx, key, 1, window).Err()
}

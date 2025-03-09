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

func (r *RedisStorage) IncrRequest(ctx context.Context, key string, window time.Duration) (int, time.Duration, error) {
	requestKey := RateLimitPrefix + "req" + key
	count := r.client.Incr(ctx, requestKey)

	r.logger.Info("Incrementing request count",
		slog.String("key", requestKey),
		slog.Int("count", int(count.Val())),
	)

	if count.Err() != nil {
		r.logger.Error("Error incrementing request count",
			slog.String("key", requestKey),
			slog.String("error", count.Err().Error()),
		)
		return 0, 0, count.Err()
	}

	if count.Val() == 1 {
		r.logger.Info("Setting expiration for key",
			slog.String("key", requestKey),
			slog.String("window", window.String()),
		)

		err := r.client.Expire(ctx, requestKey, window)
		if err.Err() != nil {
			r.logger.Error("Error setting expiration",
				slog.String("key", requestKey),
				slog.String("error", err.Err().Error()),
			)
			return int(count.Val()), 0, err.Err()
		}
	}

	ttl := r.client.TTL(ctx, requestKey)
	if ttl.Err() != nil {
		r.logger.Error("Error getting TTL",
			slog.String("key", requestKey),
			slog.String("error", ttl.Err().Error()),
		)
		return int(count.Val()), 0, ttl.Err()
	}

	return int(count.Val()), ttl.Val(), nil
}

func (r *RedisStorage) IsBlocked(ctx context.Context, key string) (bool, time.Duration, error) {
	blockKey := RateLimitPrefix + "block:" + key

	r.logger.Info("Checking if key is blocked",
		slog.String("key", blockKey),
	)

	ttl := r.client.TTL(ctx, blockKey)
	if ttl.Err() != nil {
		if ttl.Err() == redis.Nil {
			return false, 0, nil
		}

		r.logger.Error("Error getting key TTL",
			slog.String("key", blockKey),
			slog.String("error", ttl.Err().Error()),
		)

		return false, 0, ttl.Err()
	}

	if ttl.Val() > 0 {
		r.logger.Info("Key is blocked",
			slog.String("key", blockKey),
			slog.String("ttl", ttl.Val().String()),
		)
		return true, ttl.Val(), nil
	}

	r.logger.Info("Key is not blocked",
		slog.String("key", blockKey),
	)

	return false, 0, nil
}

func (r *RedisStorage) BlockRequest(ctx context.Context, key string, duration time.Duration) error {
	blockKey := RateLimitPrefix + "block:" + key

	r.logger.Info("Blocking key",
		slog.String("key", blockKey),
		slog.String("duration", duration.String()),
	)

	statusCmd := r.client.Set(ctx, blockKey, "blocked", duration)
	if statusCmd.Err() != nil {
		r.logger.Error("Error setting key expiration",
			slog.String("key", blockKey),
			slog.String("error", statusCmd.Err().Error()),
		)
		return statusCmd.Err()
	}

	return nil
}

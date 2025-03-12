package ratelimiter

import (
	"context"
	"os"
	"testing"
	"time"

	"log/slog"

	"github.com/go-redis/redismock/v9"
	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/mocks"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStorage_IncrRequest(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	logger := &mocks.LoggerMock{}
	logger.On("NewLogger").Return(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	storage := NewRedisStorage(client, logger.NewLogger())

	t.Run("increments request count and sets expiration", func(t *testing.T) {
		key := "test_key"
		window := 10 * time.Second
		requestKey := RateLimitPrefix + "req:" + key

		mock.ExpectIncr(requestKey).SetVal(1)
		mock.ExpectExpire(requestKey, window).SetVal(true)
		mock.ExpectTTL(requestKey).SetVal(window)

		count, ttl, err := storage.IncrRequest(ctx, key, window)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, window, ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("increments request count without setting expiration", func(t *testing.T) {
		key := "test_key"
		window := 10 * time.Second
		requestKey := RateLimitPrefix + "req:" + key

		mock.ExpectIncr(requestKey).SetVal(2)
		mock.ExpectTTL(requestKey).SetVal(window)

		count, ttl, err := storage.IncrRequest(ctx, key, window)
		require.NoError(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, window, ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when incrementing request count fails", func(t *testing.T) {
		key := "test_key"
		window := 10 * time.Second
		requestKey := RateLimitPrefix + "req:" + key

		mock.ExpectIncr(requestKey).SetErr(redis.Nil)

		count, ttl, err := storage.IncrRequest(ctx, key, window)
		assert.Error(t, err)
		assert.Equal(t, 0, count)
		assert.Equal(t, time.Duration(0), ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when setting expiration fails", func(t *testing.T) {
		key := "test_key"
		window := 10 * time.Second
		requestKey := RateLimitPrefix + "req:" + key

		mock.ExpectIncr(requestKey).SetVal(1)
		mock.ExpectExpire(requestKey, window).SetErr(redis.Nil)

		count, ttl, err := storage.IncrRequest(ctx, key, window)
		assert.Error(t, err)
		assert.Equal(t, 1, count)
		assert.Equal(t, time.Duration(0), ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when getting TTL fails", func(t *testing.T) {
		key := "test_key"
		window := 10 * time.Second
		requestKey := RateLimitPrefix + "req:" + key

		mock.ExpectIncr(requestKey).SetVal(2)
		mock.ExpectTTL(requestKey).SetErr(redis.Nil)

		count, ttl, err := storage.IncrRequest(ctx, key, window)
		assert.Error(t, err)
		assert.Equal(t, 2, count)
		assert.Equal(t, time.Duration(0), ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisStorage_IsBlocked(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	logger := &mocks.LoggerMock{}
	logger.On("NewLogger").Return(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	storage := NewRedisStorage(client, logger.NewLogger())

	t.Run("returns false when key is not blocked", func(t *testing.T) {
		key := "test_key"
		blockKey := RateLimitPrefix + "block:" + key

		mock.ExpectTTL(blockKey).SetErr(redis.Nil)

		blocked, ttl, err := storage.IsBlocked(ctx, key)
		require.NoError(t, err)
		assert.False(t, blocked)
		assert.Equal(t, time.Duration(0), ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns true when key is blocked", func(t *testing.T) {
		key := "test_key"
		blockKey := RateLimitPrefix + "block:" + key
		ttlDuration := 10 * time.Second

		mock.ExpectTTL(blockKey).SetVal(ttlDuration)

		blocked, ttl, err := storage.IsBlocked(ctx, key)
		require.NoError(t, err)
		assert.True(t, blocked)
		assert.Equal(t, ttlDuration, ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when getting TTL fails", func(t *testing.T) {
		key := "test_key"
		blockKey := RateLimitPrefix + "block:" + key

		mock.ExpectTTL(blockKey).SetErr(redis.ErrClosed)

		blocked, ttl, err := storage.IsBlocked(ctx, key)
		assert.Error(t, err)
		assert.False(t, blocked)
		assert.Equal(t, time.Duration(0), ttl)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisStorage_BlockRequest(t *testing.T) {
	ctx := context.Background()
	client, mock := redismock.NewClientMock()
	logger := &mocks.LoggerMock{}
	logger.On("NewLogger").Return(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	storage := NewRedisStorage(client, logger.NewLogger())

	t.Run("blocks request successfully", func(t *testing.T) {
		key := "test_key"
		duration := 10 * time.Second
		blockKey := RateLimitPrefix + "block:" + key

		mock.ExpectSet(blockKey, "blocked", duration).SetVal("OK")

		err := storage.BlockRequest(ctx, key, duration)
		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("returns error when setting key expiration fails", func(t *testing.T) {
		key := "test_key"
		duration := 10 * time.Second
		blockKey := RateLimitPrefix + "block:" + key

		mock.ExpectSet(blockKey, "blocked", duration).SetErr(redis.ErrClosed)

		err := storage.BlockRequest(ctx, key, duration)
		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

package ratelimiter

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/jhonasalves/go-expert-fc-rate-limiter/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRateLimiter_Allow(t *testing.T) {
	mockStorage := new(mocks.StorageMock)
	logger := &mocks.LoggerMock{}
	logger.On("NewLogger").Return(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	opts := Options{
		MaxRequestIP:    5,
		MaxRequestToken: 10,
		WindowDuration:  time.Minute,
		BlockDuration:   time.Minute * 5,
	}
	rateLimiter := NewRateLimiter(mockStorage, opts, logger.NewLogger())

	t.Run("should allow request when not blocked and under limit", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(1, time.Minute, nil)

		resp, err := rateLimiter.Allow(ctx, rk)

		assert.NoError(t, err)
		assert.True(t, resp.Allowed)
		assert.Equal(t, 9, resp.RequestsLeft)
		assert.Equal(t, opts.MaxRequestToken, resp.Limit)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should block request when already blocked", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(true, time.Minute, nil)

		resp, err := rateLimiter.Allow(ctx, rk)

		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Equal(t, 0, resp.RequestsLeft)
		assert.Equal(t, opts.MaxRequestToken, resp.Limit)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should block request when over limit", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(11, time.Minute, nil)
		mockStorage.On("BlockRequest", ctx, rk.Key, opts.BlockDuration).Return(nil)

		resp, err := rateLimiter.Allow(ctx, rk)

		assert.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Equal(t, 0, resp.RequestsLeft)
		assert.Equal(t, opts.MaxRequestToken, resp.Limit)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should return error when IsBlocked fails", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), errors.New("storage error"))

		resp, err := rateLimiter.Allow(ctx, rk)

		assert.Error(t, err)
		assert.False(t, resp.Allowed)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should return error when IncrRequest fails", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", ctx, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", ctx, rk.Key, opts.WindowDuration).Return(0, time.Duration(0), errors.New("storage error"))

		resp, err := rateLimiter.Allow(ctx, rk)

		assert.Error(t, err)
		assert.False(t, resp.Allowed)
		mockStorage.AssertExpectations(t)
	})

	t.Run("should handle concurrent requests correctly", func(t *testing.T) {
		defer mockStorage.ClearMocks()

		ctx := context.Background()
		rk := RateLimitKey{Key: "test-key", KeyType: Token}

		mockStorage.On("IsBlocked", mock.Anything, rk.Key).Return(false, time.Duration(0), nil)
		mockStorage.On("IncrRequest", mock.Anything, rk.Key, opts.WindowDuration).Return(1, time.Minute, nil).Times(10)

		concurrentRequests := 10
		results := make(chan RateLimiterResponse, concurrentRequests)
		errors := make(chan error, concurrentRequests)

		for range concurrentRequests {
			go func() {
				resp, err := rateLimiter.Allow(ctx, rk)
				results <- resp
				errors <- err
			}()
		}

		for range concurrentRequests {
			resp := <-results
			err := <-errors

			assert.NoError(t, err)
			assert.True(t, resp.Allowed)
			assert.Equal(t, opts.MaxRequestToken-1, resp.RequestsLeft)
			assert.Equal(t, opts.MaxRequestToken, resp.Limit)
		}

		mockStorage.AssertExpectations(t)
	})
}
